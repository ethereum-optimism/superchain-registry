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
	t.Run(GenesisHashTest, func(t *testing.T) { testGenesisHash(t, chain.ChainID) })
	t.Run(GenesisRPCTest, func(t *testing.T) { testGenesisHashAgainstRPC(t, chain) })
	t.Run(UniquenessTest, func(t *testing.T) { testIsGloballyUnique(t, chain) })
	t.Run(ChainIDRPCTest, func(t *testing.T) { testChainIDFromRPC(t, chain) })
	t.Run(OptimismConfigTest, func(t *testing.T) { testOptimismConfig(t, chain) })
}

// testStandardCandidate applies to Standard and Standard Candidate Chains.
func testStandardCandidate(t *testing.T, chain *ChainConfig) {
	// Standard Config Params
	t.Run(RollupConfigTest, func(t *testing.T) { testRollupConfig(t, chain) })
	t.Run(GasTokenTest, func(t *testing.T) { testGasToken(t, chain) })
	t.Run(ResourceConfigTest, func(t *testing.T) { testResourceConfig(t, chain) })
	t.Run(GasLimitTest, func(t *testing.T) { testGasLimit(t, chain) })
	t.Run(GPOParamsTest, func(t *testing.T) { testGasPriceOracleParams(t, chain) })
	t.Run(StartBlockRPCTest, func(t *testing.T) { testStartBlock(t, chain) })
	// Standard Config Roles
	t.Run(L2SecurityConfigTest, func(t *testing.T) { testL2SecurityConfig(t, chain) })
	// Other
	t.Run(DataAvailabilityTypeTest, func(t *testing.T) { testDataAvailabilityType(t, chain) })
	t.Run(GenesisAllocsMetadataTest, func(t *testing.T) { testGenesisAllocsMetadata(t, chain) })
}

// testStandard should be applied only to a fully Standard Chain,
// i.e. not to a Standard Candidate Chain.
func testStandard(t *testing.T, chain *ChainConfig) {
	// Standard Config Params
	t.Run(SuperchainConfigTest, func(t *testing.T) { testSuperchainConfig(t, chain) })
	// Standard Contract Versions
	t.Run(StandardContractVersionsTest, func(t *testing.T) { testContractsMatchATag(t, chain) })
	// Standard Config Roles
	t.Run(L1SecurityConfigTest, func(t *testing.T) { testL1SecurityConfig(t, chain) })
	// Standard Config Params
	t.Run(OptimismPortal2ParamsTest, func(t *testing.T) { testOptimismPortal2Params(t, chain) })
	// Standard Config Roles
	t.Run(KeyHandoverTest, func(t *testing.T) { testKeyHandover(t, chain.ChainID) })
}
