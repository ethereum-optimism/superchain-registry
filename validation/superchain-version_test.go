package validation

import (
	"testing"

	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/stretchr/testify/require"
)

func TestCheckMatchOrTestnet(t *testing.T) {
	standardVersions := standard.NetworkVersions["mainnet"].Releases[standard.Release]
	dummyVersions := standard.ContractVersions{
		OptimismPortal: standard.VersionedContract{Version: "incorrect"},
		SystemConfig:   standard.VersionedContract{ImplementationAddress: nil},
	}

	matches := checkMatchOrTestnet(standardVersions, dummyVersions, false)
	require.False(t, matches)
}
