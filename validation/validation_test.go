package validation

import (
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/common"
)

// Test names
const (
	GenesisHashTest              = "Genesis_Hash"
	GenesisRPCTest               = "Genesis_RPC"
	UniquenessTest               = "Uniqueness"
	ChainIDRPCTest               = "ChainID_RPC"
	OptimismConfigTest           = "Optimism_Config"
	RollupConfigTest             = "Rollup_Config"
	GasTokenTest                 = "Gas_Token"
	ResourceConfigTest           = "Resource_Config"
	GasLimitTest                 = "Gas_Limit"
	GPOParamsTest                = "GPO_Params"
	StartBlockRPCTest            = "Start_Block_RPC"
	SuperchainConfigTest         = "Superchain_Config"
	L1SecurityConfigTest         = "L1_Security_Config"
	L2SecurityConfigTest         = "L2_Security_Config"
	DataAvailabilityTypeTest     = "Data_Availability_Type"
	StandardContractVersionsTest = "Standard_Contract_Versions"
	OptimismPortal2ParamsTest    = "Optimism_Portal_2_Params"
	KeyHandoverTest              = "Key_Handover"
	GenesisAllocsMetadataTest    = "Genesis_Allocs_Metadata"
)

type (
	subTest         = func(t *testing.T)
	subTestForChain = func(t *testing.T, chain *ChainConfig)
)

// applyExclusions is a higher order function which returns a subtest function with exclusions applied
func applyExclusions(chain *ChainConfig, f subTestForChain) subTest {
	return func(t *testing.T) {
		skipIfExcluded(t, chain.ChainID)
		f(t, chain)
	}
}

func TestValidation(t *testing.T) {
	// Entry point for validation checks which run
	// on each OP chain.
	for _, chain := range OPChains {
		chain := chain
		t.Run(common.PerChainTestName(chain), func(t *testing.T) {
			t.Parallel()
			testValidation(t, chain)
		})
	}
}

func testValidation(t *testing.T, chain *ChainConfig) {
	testUniversal(t, chain)

	if chain.SuperchainLevel == Standard ||
		(chain.SuperchainLevel == Frontier && chain.StandardChainCandidate) {
		testStandardCandidate(t, chain)
	}

	if chain.SuperchainLevel == Standard {
		testStandard(t, chain)
	}
}

// testUniversal should be applied to each chain in the registry
// regardless of superchain_level. There should be no
// exceptions, since the test is e.g.
// designed to protect downstream software or
// sanity checking basic consistency conditions.
func testUniversal(t *testing.T, chain *ChainConfig) {
	t.Run(GenesisHashTest, applyExclusions(chain, testGenesisHash))
	t.Run(GenesisRPCTest, applyExclusions(chain, testGenesisHashAgainstRPC))
	t.Run(UniquenessTest, applyExclusions(chain, testIsGloballyUnique))
	t.Run(ChainIDRPCTest, applyExclusions(chain, testChainIDFromRPC))
	t.Run(OptimismConfigTest, applyExclusions(chain, testOptimismConfig))
}

// testStandardCandidate applies to Standard and Standard Candidate Chains.
func testStandardCandidate(t *testing.T, chain *ChainConfig) {
	// Standard Config Params
	t.Run(RollupConfigTest, applyExclusions(chain, testRollupConfig))
	t.Run(GasTokenTest, applyExclusions(chain, testGasToken))
	t.Run(ResourceConfigTest, applyExclusions(chain, testResourceConfig))
	t.Run(GasLimitTest, applyExclusions(chain, testGasLimit))
	t.Run(GPOParamsTest, applyExclusions(chain, testGasPriceOracleParams))
	t.Run(StartBlockRPCTest, applyExclusions(chain, testStartBlock))
	// Standard Config Roles
	t.Run(L2SecurityConfigTest, applyExclusions(chain, testL2SecurityConfig))
	// Other
	t.Run(DataAvailabilityTypeTest, applyExclusions(chain, testDataAvailabilityType))
	t.Run(GenesisAllocsMetadataTest, applyExclusions(chain, testGenesisAllocsMetadata))
}

// testStandard should be applied only to a fully Standard Chain,
// i.e. not to a Standard Candidate Chain.
func testStandard(t *testing.T, chain *ChainConfig) {
	// Standard Config Params
	t.Run(SuperchainConfigTest, applyExclusions(chain, testSuperchainConfig))
	// Standard Contract Versions
	t.Run(StandardContractVersionsTest, applyExclusions(chain, testContractsMatchATag))
	// Standard Config Roles
	t.Run(L1SecurityConfigTest, applyExclusions(chain, testL1SecurityConfig))
	// Standard Config Params
	t.Run(OptimismPortal2ParamsTest, applyExclusions(chain, testOptimismPortal2Params))
	// Standard Config Roles
	t.Run(KeyHandoverTest, applyExclusions(chain, testKeyHandover))
}
