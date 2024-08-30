package common

import (
	"fmt"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

// PerChainTestName ensures test can easily be filtered by chain name or chain id using the -run=regex testflag.
func PerChainTestName(chain *ChainConfig) string {
	return chain.Name + fmt.Sprintf(" (%d)", chain.ChainID)
}
