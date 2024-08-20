package validation

import (
	"regexp"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
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
	}
}

var exclusions = map[string]map[uint64]bool{
	// Universal Checks
	"Genesis_Hash_Check": {
		// OP Mainnet has a pre-bedrock genesis (with an empty allocs object stored in the registry), so we exclude it from this check.")
		10: true,
	},
	"ChainID_RPC_Check": {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	},
	"Genesis_RPC_Check": {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	},
	"Uniqueness_Check": {
		11155421: true, // oplabs devnet 0, not in upstream repo
		11763072: true, // base devnet 0, not in upstream repo}
	},

	// Standard Checks
	"L1_Security_Config": {
		8453:      true, // base (incorrect challenger, incorrect guardian)
		84532:     true, // base-sepolia (incorrect challenger)
		7777777:   true, // zora (incorrect challenger)
		1750:      true, // metal (incorrect challenger)
		919:       true, // mode sepolia (incorrect challenger)
		999999999: true, // zora sepolia (incorrect challenger)
		34443:     true, // mode (incorrect challenger)
		1740:      true, // metal-sepolia
	},
	"Standard_Contract_Versions": {
		8453:      true, // mainnet/base         MCP (at time of writing)
		1750:      true, // mainnet/metal        MCP (at time of writing)
		34443:     true, // mainnet/mode         MCP (at time of writing)
		7777777:   true, // mainnet/zora         MCP (at time of writing)
		84532:     true, // sepolia/base         MCP (at time of writing)
		1740:      true, // sepolia/metal        MCP (at time of writing)
		919:       true, // sepolia/mode         MCP (at time of writing)
		999999999: true, // sepolia/zora         MCP (at time of writing)
		11155421:  true, // sepolia-dev0/oplabs-devnet-0
		11763072:  true, // sepolia-dev0/base-devnet-0
	},
	"Optimism_Portal_2_Params": {
		8453:      true, // mainnet/base         MCP (at time of writing)
		1750:      true, // mainnet/metal        MCP (at time of writing)
		34443:     true, // mainnet/mode         MCP (at time of writing)
		7777777:   true, // mainnet/zora         MCP (at time of writing)
		84532:     true, // sepolia/base         MCP (at time of writing)
		1740:      true, // sepolia/metal        MCP (at time of writing)
		919:       true, // sepolia/mode         MCP (at time of writing)
		999999999: true, // sepolia/zora         MCP (at time of writing)
		11763072:  true, // sepolia-dev0/base-devnet-0
	},
}

func TestExclusions(t *testing.T) {
	for name, v := range exclusions {
		for k := range v {
			if k == 10 && name == "Genesis_Hash_Check" {
				// This is the sole standard chain validation check exclusion
				continue
			}
			if v[k] {
				require.NotNil(t, superchain.OPChains[k], k)
				require.False(t, superchain.OPChains[k].SuperchainLevel == superchain.Standard, "Standard Chain %d may not be excluded from any check", k)
			}
		}
	}
}
