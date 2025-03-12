package validation

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/stretchr/testify/require"
)

// This test file tests the integrity of the standard versions files. However, it can't
// live in the validation package because the validation package can't import Geth.

var (
	versionFn = w3.MustNewFunc("version()", "string")
	implsFn   = w3.MustNewFunc("implementations()", "(address superchainConfig,address protocolVersions,address l1ERC721Bridge,address optimismPortal,address systemConfig,address optimismMintableERC20Factory,address l1CrossDomainMessenger,address l1StandardBridge,address disputeGameFactory,address anchorStateRegistry,address delayedWeth,address mips)")
	oracleFn  = w3.MustNewFunc("oracle()", "address")
)

type opcmImpls struct {
	SuperchainConfig             common.Address
	ProtocolVersions             common.Address
	L1ERC721Bridge               common.Address
	OptimismPortal               common.Address
	SystemConfig                 common.Address
	OptimismMintableERC20Factory common.Address
	L1CrossDomainMessenger       common.Address
	L1StandardBridge             common.Address
	DisputeGameFactory           common.Address
	AnchorStateRegistry          common.Address
	DelayedWeth                  common.Address
	Mips                         common.Address
}

var rpcURLs = map[string]string{
	"sepolia": os.Getenv("SEPOLIA_RPC_URL"),
	"mainnet": os.Getenv("MAINNET_RPC_URL"),
}

var versionMappings = map[string]validation.Versions{
	"sepolia": validation.StandardVersionsSepolia,
	"mainnet": validation.StandardVersionsMainnet,
}

var versionsToCheck = []validation.Semver{
	"op-contracts/v2.0.0-rc.1",
	"op-contracts/v3.0.0-rc.1",
}

func TestVersionsIntegrity(t *testing.T) {
	for _, network := range []string{"sepolia", "mainnet"} {
		t.Run(network, func(t *testing.T) {
			rpcURL := rpcURLs[network]
			require.NotEmpty(t, rpcURL)

			versions := versionMappings[network]
			require.NotEmpty(t, versions)

			rpcClient, err := rpc.Dial(rpcURL)
			require.NoError(t, err)

			w3Client := w3.NewClient(rpcClient)

			for _, vc := range versionsToCheck {
				t.Run(string(vc), func(t *testing.T) {
					testVersionIntegrity(t, versions[vc], w3Client)
				})
			}
		})
	}
}

func testVersionIntegrity(t *testing.T, stdVer validation.VersionConfig, w3Client *w3.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	require.NotNil(t, stdVer.OPContractsManager)
	opcmAddr := stdVer.OPContractsManager.Address
	require.NotNil(t, opcmAddr)

	var impls opcmImpls
	require.NoError(t, w3Client.CallCtx(
		ctx,
		eth.CallFunc(common.Address(*opcmAddr), implsFn).Returns(&impls),
	))

	vValue := reflect.ValueOf(&stdVer).Elem()
	implsValue := reflect.ValueOf(impls)

	fields := []string{
		"SuperchainConfig",
		"ProtocolVersions",
		"L1ERC721Bridge",
		"OptimismPortal",
		"SystemConfig",
		"OptimismMintableERC20Factory",
		"L1CrossDomainMessenger",
		"L1StandardBridge",
		"DisputeGameFactory",
		"AnchorStateRegistry",
		"DelayedWeth",
		"Mips",
	}

	for _, field := range fields {
		implsField := implsValue.FieldByName(field)
		require.True(t, implsField.IsValid())

		address := implsField.Interface().(common.Address)
		contractData := vValue.FieldByName(field).Interface().(*validation.ContractData)
		require.NotNil(t, contractData)

		if contractData.Address != nil {
			require.Equal(t, common.Address(*contractData.Address), address, "invalid address for %s", field)
		} else if contractData.ImplementationAddress != nil {
			require.Equal(t, common.Address(*contractData.ImplementationAddress), address, "invalid implementation address for %s", field)
		} else {
			require.Empty(t, address, "address %s should be empty", field)
		}

		var contractVer string
		require.NoError(t, w3Client.CallCtx(ctx, eth.CallFunc(address, versionFn).Returns(&contractVer)))
		require.Equal(t, contractData.Version, contractVer, "invalid version for %s", field)
	}

	var oracleAddr common.Address
	require.NoError(t, w3Client.CallCtx(ctx, eth.CallFunc(common.Address(*stdVer.Mips.Address), oracleFn).Returns(&oracleAddr)))
	require.Equal(t, common.Address(*stdVer.PreimageOracle.Address), oracleAddr, "invalid oracle address")
}
