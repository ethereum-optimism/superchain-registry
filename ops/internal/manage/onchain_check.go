package manage

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
)

// OnchainMismatch describes a single field in a chain's TOML config that
// disagrees with the value fetched from onchain.
type OnchainMismatch struct {
	Section string // TOML table, e.g. "roles" or "addresses"
	Field   string // TOML key, e.g. "ProxyAdminOwner"
	Disk    config.ChecksummedAddress
	Onchain config.ChecksummedAddress
}

// Unresolved reports whether the onchain value is the zero address, i.e. the
// fetcher could not resolve the field onchain (e.g. no DisputeGameFactory is
// wired into a chain that has not adopted fault proofs). This is not proof
// that the TOML is wrong, so callers should treat it as a warning.
func (m OnchainMismatch) Unresolved() bool {
	return m.Onchain == config.ChecksummedAddress{}
}

func (m OnchainMismatch) String() string {
	return fmt.Sprintf("[%s] %s: toml=%s onchain=%s", m.Section, m.Field, m.Disk, m.Onchain)
}

// DiffChainAgainstOnchain compares the onchain-derived fields ([roles] and
// [addresses]) of a chain's TOML config against freshly fetched onchain data.
// Only fields that are present in the TOML are compared, since many chain
// configs intentionally list only a subset of roles/addresses.
func DiffChainAgainstOnchain(disk *config.Chain, onchain config.AddressesWithRoles) []OnchainMismatch {
	var out []OnchainMismatch
	out = append(out, diffAddressStructs("roles", disk.Roles, onchain.Roles)...)
	out = append(out, diffAddressStructs("addresses", disk.Addresses, onchain.Addresses)...)
	return out
}

// diffAddressStructs compares two structs of the same type whose fields are all
// *config.ChecksummedAddress. Fields that are nil on the disk side are skipped.
func diffAddressStructs(section string, disk, onchain any) []OnchainMismatch {
	dv := reflect.ValueOf(disk)
	ov := reflect.ValueOf(onchain)
	if dv.Type() != ov.Type() {
		panic(fmt.Sprintf("type mismatch: %s != %s", dv.Type(), ov.Type()))
	}

	addrPtrType := reflect.TypeOf((*config.ChecksummedAddress)(nil))
	var out []OnchainMismatch
	for i := 0; i < dv.NumField(); i++ {
		field := dv.Type().Field(i)
		if field.Type != addrPtrType {
			continue
		}
		diskAddr := dv.Field(i).Interface().(*config.ChecksummedAddress)
		if diskAddr == nil {
			continue
		}
		var onchainAddr config.ChecksummedAddress
		if p := ov.Field(i).Interface().(*config.ChecksummedAddress); p != nil {
			onchainAddr = *p
		}
		if *diskAddr != onchainAddr {
			out = append(out, OnchainMismatch{
				Section: section,
				Field:   tomlFieldName(field),
				Disk:    *diskAddr,
				Onchain: onchainAddr,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Field < out[j].Field })
	return out
}

func tomlFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("toml")
	if name, _, _ := strings.Cut(tag, ","); name != "" {
		return name
	}
	return f.Name
}
