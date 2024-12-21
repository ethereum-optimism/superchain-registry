package common

import (
	"fmt"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

// PerChainTestName formats a test name with the chain name and chain ID,
// allowing tests to be filtered using the -run=regex test flag.
func PerChainTestName(chain *ChainConfig) string {
	return chain.Name + fmt.Sprintf(" (%d)", chain.ChainID)
}
