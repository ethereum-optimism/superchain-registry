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

var exclusions = map[string]map[uint64]bool{
	// Universal Checks
	GenesisHashTest: {
		// OP Mainnet has a pre-bedrock genesis (with an empty allocs object stored in the registry), so we exclude it from this check.")
		10: true,
	},
	ChainIDRPCTest: {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	},
	GenesisRPCTest: {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	},
	UniquenessTest: {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   Not in https://github.com/ethereum-lists/chains
		11763072: true, // sepolia-dev-0/base-devnet-0     Not in https://github.com/ethereum-lists/chains
	},

	// Standard Checks
	OptimismPortal2ParamsTest: {
		11763072: true, // sepolia-dev0/base-devnet-0
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
			if k == 10 && name == GenesisHashTest {
				// This is the sole standard chain validation check exclusion
				continue
			}
			if v[k] {
				require.NotNil(t, OPChains[k], k)
				require.False(t, OPChains[k].SuperchainLevel == Standard, "Standard Chain %d may not be excluded from any check", k)
			}
		}
	}
}
