package validation

import (
	"reflect"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"
)

func TestCheckMatchOrTestnet(t *testing.T) {
	dummyVersions := standard.ContractVersions{
		OptimismPortal: standard.VersionedContract{Version: "incorrect"},
		SystemConfig:   standard.VersionedContract{ImplementationAddress: nil},
	}

	standardVersions := standard.NetworkVersions["mainnet"].Releases[standard.Release]

	s := reflect.ValueOf(standardVersions)
	c := reflect.ValueOf(dummyVersions)

	matches := checkMatchOrTestnet(s, c, false)
	require.False(t, matches)
}
