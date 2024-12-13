package validation

import (
	"context"
	"math/big"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/common"
	"github.com/stretchr/testify/require"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Test names
const (
	GenesisHashTest              = "Genesis_Hash"
	GenesisRPCTest               = "Genesis_RPC"
	UniquenessTest               = "Uniqueness"
	PublicRPCTest                = "Public_RPC"
	ChainIDRPCTest               = "ChainID_RPC"
	OptimismConfigTest           = "Optimism_Config"
	GovernedByOptimismTest       = "Governed_By_Optimism"
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
	FaultGameParamsTest          = "Fault_Game_Params"
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

func preflightChecks(t *testing.T) {
	// Check that all superchains have an accessible L1 archive RPC endpoint configured
	for name, chain := range Superchains {
		rpcEndpoint := chain.Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint, "no public_rpc specified for superchain '%s'", name)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint '%s' for superchain '%s'", rpcEndpoint, name)
		defer client.Close()

		_, err = client.ChainID(context.Background())
		require.NoErrorf(t, err, "could not query node at '%s' for superchain '%s'", rpcEndpoint, name)

		superchainConfigAddr := *chain.Config.SuperchainConfigAddr

		_, err = client.NonceAt(context.Background(), ethCommon.Address(superchainConfigAddr), big.NewInt(1))
		require.NoErrorf(t, err, "node at '%s' for superchain '%s' is not an archive node", rpcEndpoint, name)
	}
}

func TestValidation(t *testing.T) {
	preflightChecks(t)

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
	t.Run(PublicRPCTest, applyExclusions(chain, testPublicRPC))
	t.Run(ChainIDRPCTest, applyExclusions(chain, testChainIDFromRPC))
	t.Run(OptimismConfigTest, applyExclusions(chain, testOptimismConfig))
	t.Run(GovernedByOptimismTest, applyExclusions(chain, testGovernedByOptimism))
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
	t.Run(SuperchainConfigTest, applyExclusions(chain, testSuperchainConfig))
	t.Run(StandardContractVersionsTest, applyExclusions(chain, checkForStandardVersions))
	t.Run(FaultGameParamsTest, applyExclusions(chain, testFaultGameParams))
	t.Run(L1SecurityConfigTest, applyExclusions(chain, testL1SecurityConfig))
	t.Run(OptimismPortal2ParamsTest, applyExclusions(chain, testOptimismPortal2Params))
	t.Run(KeyHandoverTest, applyExclusions(chain, testKeyHandover))
}
