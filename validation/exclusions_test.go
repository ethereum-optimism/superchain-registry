package validation

import (
	"regexp"
	"testing"
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
	"ChainID_RPC_Check": {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	},
	"GPO_Params": {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   (no public endpoint)
		11763072: true, // sepolia-dev-0/base-devnet-0     (no public endpoint)
	},
	"Superchain_Config": {
		11763072: true, // sepolia-dev-0/base-devnet-0 (old version of OptimismPortal)
	},
	"L2OO_Params": {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0 (does not yet declare a contract versions tag)
		11763072: true, // sepolia-dev-0/base-devnet-0  (does not yet declare a contract versions tag)
	},
	"L1_Security_Config": {
		8453:      true, // base (incorrect challenger, incorrect guardian)
		11763072:  true, // base-devnet-0 (incorrect challenger, incorrect guardian)
		90001:     true, // race (incorrect challenger, incorrect guardian)
		84532:     true, // base-sepolia (incorrect challenger)
		7777777:   true, // zora (incorrect challenger)
		1750:      true, // metal (incorrect challenger)
		919:       true, // mode sepolia (incorrect challenger)
		999999999: true, // zora sepolia (incorrect challenger)
		34443:     true, // mode (incorrect challenger)
	},
	"L2_Security_Config": {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	},
	"Genesis_Hash_Check": {
		10: true, // OP Mainnet
	},
	"Genesis_RPC_Check": {
		11155421: true, // sepolia-dev-0/oplabs-devnet-0   No Public RPC declared
		11763072: true, // sepolia-dev-0/base-devnet-0     No Public RPC declared
	},
	"Standard_Contract_Versions": {
		10:        true, // mainnet/op           MCP (at time of writing)
		8453:      true, // mainnet/base         MCP (at time of writing)
		957:       true, // mainnet/lyra         MCP (at time of writing)
		1750:      true, // mainnet/metal        MCP (at time of writing)
		34443:     true, // mainnet/mode         MCP (at time of writing)
		291:       true, // mainnet/orderly      MCP (at time of writing)
		424:       true, // mainnet/pgn          MCP (at time of writing)
		7777777:   true, // mainnet/zora         MCP (at time of writing)
		11155420:  true, // sepolia/op           MCP (at time of writing)
		84532:     true, // sepolia/base         MCP (at time of writing)
		919:       true, // sepolia/mode         MCP (at time of writing)
		58008:     true, // sepolia/pgn          MCP (at time of writing)
		999999999: true, // sepolia/zora         MCP (at time of writing)
		11155421:  true, // sepolia-dev0/oplabs-devnet-0"
		11763072:  true, // sepolia-dev0/base-devnet-0"
	},
	"Uniqueness_Check": {
		11155421: true, // oplabs devnet 0, not in upstream repo
		11763072: true, // base devnet 0, not in upstream repo}
	},
}
