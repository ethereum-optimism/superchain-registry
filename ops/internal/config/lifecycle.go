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

// hardforkModifiers maps each non-timestamp field of [Hardforks] to the hardfork
// whose activation governs when it freezes. Such a field is not itself an activation
// but records how its hardfork behaved when it activated (e.g. keep_karst_upgrade_gas
// records whether a chain kept the inflated gas limit that Karst introduced). It may
// therefore be set or changed freely until its governing hardfork is in the past,
// after which the fact is on-chain history and frozen. New non-timestamp fields added
// to [Hardforks] must be registered here (TestHardforksFieldsClassified enforces it).
var hardforkModifiers = map[string]string{
	"KeepKarstUpgradeGas": "KarstTime",
}

// checkAppendOnly verifies the append-only contract for each entry of a struct field
// (e.g. the activation times in [Hardforks]). Each entry is one of:
//
//   - A hardfork activation (*HardforkTime). An unset activation may be added later; a
//     set activation still in the future may be re-scheduled or removed; once its time
//     is at or before now it has activated on-chain and is frozen forever.
//   - A non-timestamp modifier recording how a hardfork behaved on activation (see
//     [hardforkModifiers]). It may be set or changed freely until its governing
//     hardfork is in the past, after which it is frozen. Because a bool has no "unset"
//     state distinct from false, its freeze is driven by the governing hardfork's
//     time rather than by the field's own value.
//
// A changed entry that is neither is a programming error — a new entry type was added
// to an append-only struct without deciding how it ages — and is reported so it can
// never slip through unchecked.
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
		of, nf := oldV.Field(i), newV.Field(i)
		field := t.Field(i)
		name := parent + "." + tomlName(field)

		switch {
		case isHardforkTime(field.Type):
			if of.IsNil() {
				continue // unset activation may be added later
			}
			if of.Elem().Uint() > now {
				continue // future activation may still be re-scheduled or removed
			}
			if !reflect.DeepEqual(of.Interface(), nf.Interface()) {
				violations = append(violations, fmt.Sprintf("%q is append-only; an activation already in the past must not change or be removed", name))
			}
		case hardforkModifiers[field.Name] != "":
			gate, gated := hardforkActivation(oldV.FieldByName(hardforkModifiers[field.Name]))
			if !gated || gate > now {
				continue // governing hardfork not yet activated: modifier may still change
			}
			if !reflect.DeepEqual(of.Interface(), nf.Interface()) {
				violations = append(violations, fmt.Sprintf("%q records how an already-activated hardfork behaved and must not change once that hardfork is in the past", name))
			}
		default:
			// Fail closed: an unclassified change must not pass silently.
			if !reflect.DeepEqual(of.Interface(), nf.Interface()) {
				violations = append(violations, fmt.Sprintf("%q changed but has no append-only rule; classify it in checkAppendOnly before merging", name))
			}
		}
	}
	return violations
}

// isHardforkTime reports whether t is *HardforkTime.
func isHardforkTime(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem() == hardforkTimeType
}

// hardforkActivation returns the activation time (Unix seconds) of a set
// *HardforkTime field, and whether the field is in fact a non-nil HardforkTime.
func hardforkActivation(v reflect.Value) (uint64, bool) {
	if !isHardforkTime(v.Type()) || v.IsNil() {
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
