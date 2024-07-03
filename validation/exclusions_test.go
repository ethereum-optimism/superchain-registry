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
	"L2OO_Params": {
		10:        true, // OP mainnet - Upgraded to fault proofs
		999999999: true, // sepolia/zora  Incorrect finalizationPeriodSeconds, 604800 is not within bounds [12 12]
		1740:      true, // sepolia/metal Incorrect finalizationPeriodSeconds, 604800 is not within bounds [12 12]
		919:       true, // sepolia/mode  Incorrect finalizationPeriodSeconds, 180 is not within bounds [12 12]
		11155420:  true, // sepolia/op No L2OO because this chain uses Fault Proofs https://github.com/ethereum-optimism/superchain-registry/issues/219
		11155421:  true, // oplabs-sepolia-devnet-0 No L2OO because this chain uses Fault Proofs https://github.com/ethereum-optimism/superchain-registry/issues/219
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
		8453:     true, // mainnet/base
		11155421: true, // sepolia-dev-0/oplabs-devnet-0
		11763072: true, // sepolia-dev-0/base-devnet-0
	}}
