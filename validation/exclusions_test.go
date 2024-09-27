package validation

import (
	"regexp"
	"testing"
	"time"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"
)

func skipIfExcluded(t *testing.T, chainID uint64) {
	for pattern := range exclusions {
		matches, err := regexp.Match(pattern, []byte(t.Name()))
		if err != nil {
			panic(err)
		}
		if matches && exclusions[pattern][chainID] {
			t.Skip("Excluded!")
		}
		if matches && silences[pattern][chainID].After(time.Now()) {
			t.Skipf("Silenced until %s", silences[pattern][chainID].String())
		}
	}
}

var opcmTestChainId = uint64(111222333444555666)

var exclusions = map[string]map[uint64]bool{
	// Universal Checks
	GenesisHashTest: {
		opcmTestChainId: true,
		// OP Mainnet has a pre-bedrock genesis (with an empty allocs object stored in the registry), so we exclude it from this check.")
		10: true,
	},
	GenesisAllocsMetadataTest: {
		opcmTestChainId: true,
		10:              true, // op-mainnet
		1740:            true, // metal-sepolia
	},
	ChainIDRPCTest: {
		opcmTestChainId: true,
		11155421:        true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072:        true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	},
	GenesisRPCTest: {
		opcmTestChainId: true,
		11155421:        true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072:        true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	},
	UniquenessTest: {
		opcmTestChainId: true,
		11155421:        true, // sepolia-dev-0/oplabs-devnet-0   Not in https://github.com/ethereum-lists/chains
		11763072:        true, // sepolia-dev-0/base-devnet-0     Not in https://github.com/ethereum-lists/chains
	},

	// Standard Checks
	OptimismPortal2ParamsTest: {
		11763072: true, // sepolia-dev0/base-devnet-0
	},

	OptimismConfigTest: {
		opcmTestChainId: true,
	},
	RollupConfigTest: {
		opcmTestChainId: true,
	},
	GasTokenTest: {
		opcmTestChainId: true,
	},
	GPOParamsTest: {
		opcmTestChainId: true,
	},
	L2SecurityConfigTest: {
		opcmTestChainId: true,
	},
	DataAvailabilityTypeTest: {
		opcmTestChainId: true,
	},
}

var silences = map[string]map[uint64]time.Time{
	OptimismPortal2ParamsTest: {
		10: time.Unix(int64(*OPChains[10].HardForkConfiguration.GraniteTime), 0), // mainnet/op silenced until Granite activates
	},
}

func TestExclusions(t *testing.T) {
	for name, v := range exclusions {
		for k := range v {
			if (k == 10 || k == 11155420 || k == opcmTestChainId) && (name == GenesisHashTest || name == GenesisAllocsMetadataTest) {
				// These are the sole standard chain validation check exclusions
				continue
			}
			if v[k] {
				require.NotNil(t, OPChains[k], k)
				require.False(t, OPChains[k].SuperchainLevel == Standard, "Standard Chain %d may not be excluded from any check", k)
			}
		}
	}
}
