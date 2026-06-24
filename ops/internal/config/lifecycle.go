package config

import (
	"fmt"
	"reflect"
	"strings"
)

// FieldLifecycle classifies how a [Chain] configuration field may change over a
// chain's lifetime. Each field of [Chain] is annotated with a `lifecycle` struct
// tag holding one of these values.
type FieldLifecycle string

const (
	// LifecycleImmutable fields are fixed at chain creation and can never change.
	// Changing one describes a different chain, so [CheckImmutableFields] rejects it.
	LifecycleImmutable FieldLifecycle = "immutable"
	// LifecycleAppendOnly fields grow over time as the protocol evolves. New entries
	// may be added, and a not-yet-active entry may still be adjusted, but an entry
	// whose activation is already in the past is frozen: once a hardfork has
	// activated on-chain its timestamp is historical fact and must not change.
	LifecycleAppendOnly FieldLifecycle = "append-only"
	// LifecycleMutable fields track live on-chain or operational state and may be
	// updated freely (role rotations, contract upgrades, RPC URLs, etc.).
	LifecycleMutable FieldLifecycle = "mutable"
)

const lifecycleTag = "lifecycle"

var hardforkTimeType = reflect.TypeOf(HardforkTime(0))

// CheckImmutableFields verifies that the immutable and append-only fields of a chain
// config have not changed in a disallowed way between a previously-committed version
// (old) and a proposed version (new). It returns an error describing every violation
// found, or nil if the change is permitted. Mutable fields are ignored entirely.
//
// now is the reference time (Unix seconds) used to decide whether an append-only
// hardfork activation is already in the past (and therefore frozen) or still in the
// future (and therefore adjustable). Callers normally pass the current wall-clock
// time; tests pass a fixed value for determinism.
//
// The classification is driven by the `lifecycle` struct tags on [Chain], so this
// function stays correct as fields are added or reclassified.
func CheckImmutableFields(old, new *Chain, now uint64) error {
	oldV := reflect.ValueOf(*old)
	newV := reflect.ValueOf(*new)
	t := oldV.Type()

	var violations []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := tomlName(field)

		switch FieldLifecycle(field.Tag.Get(lifecycleTag)) {
		case LifecycleImmutable:
			if !reflect.DeepEqual(oldV.Field(i).Interface(), newV.Field(i).Interface()) {
				violations = append(violations, fmt.Sprintf("%q is immutable and must not change once a chain is registered", name))
			}
		case LifecycleAppendOnly:
			violations = append(violations, checkAppendOnly(name, oldV.Field(i), newV.Field(i), now)...)
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("disallowed change to a chain config:\n  - %s", strings.Join(violations, "\n  - "))
	}
	return nil
}

// checkAppendOnly verifies the append-only contract for each entry of a struct field
// (e.g. the activation times in [Hardforks]). For each entry that was already set in
// the old version:
//   - a hardfork activation whose time is at or before now is frozen and must be
//     unchanged (it has activated on-chain and is historical fact);
//   - a hardfork activation still in the future may be adjusted or removed;
//   - any other already-set entry type is frozen (strict append-only).
//
// Entries that were unset (nil/zero) in the old version may be freely added.
func checkAppendOnly(parent string, oldV, newV reflect.Value, now uint64) []string {
	if oldV.Kind() == reflect.Ptr {
		if oldV.IsNil() {
			return nil
		}
		if newV.IsNil() {
			return []string{fmt.Sprintf("%q is append-only and must not be removed", parent)}
		}
		oldV, newV = oldV.Elem(), newV.Elem()
	}
	if oldV.Kind() != reflect.Struct {
		return nil
	}

	var violations []string
	t := oldV.Type()
	for i := 0; i < t.NumField(); i++ {
		of := oldV.Field(i)
		if isNilOrZero(of) {
			continue // unset entries may be added later
		}
		name := parent + "." + tomlName(t.Field(i))

		// A hardfork activation that is still in the future may be re-scheduled;
		// only once it is in the past is it frozen.
		if activation, ok := hardforkActivation(of); ok && activation > now {
			continue
		}

		if !reflect.DeepEqual(of.Interface(), newV.Field(i).Interface()) {
			violations = append(violations, fmt.Sprintf("%q is append-only; an activation already in the past must not change or be removed", name))
		}
	}
	return violations
}

// hardforkActivation returns the activation time (Unix seconds) of a set
// *HardforkTime field, and whether the field is in fact a non-nil HardforkTime.
func hardforkActivation(v reflect.Value) (uint64, bool) {
	if v.Kind() != reflect.Ptr || v.IsNil() || v.Type().Elem() != hardforkTimeType {
		return 0, false
	}
	return v.Elem().Uint(), true
}

// tomlName returns the TOML key for a struct field, falling back to the Go field
// name when no toml tag is present.
func tomlName(field reflect.StructField) string {
	tag := field.Tag.Get("toml")
	if tag == "" || tag == "-" {
		return field.Name
	}
	return strings.Split(tag, ",")[0]
}

// isNilOrZero reports whether a value is a nil pointer or the zero value of its type.
func isNilOrZero(v reflect.Value) bool {
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return v.IsZero()
}
