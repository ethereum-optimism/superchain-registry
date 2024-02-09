package validation

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var isSemverAcceptable = func(desired, actual string) bool {
	return desired == actual
}

func TestSuperchainWideContractVersions(t *testing.T) {

	checkSuperchainTargetSatisfiesSemver := func(t *testing.T, superchain *Superchain) {
		rpcEndpoint := superchain.Config.L1.PublicRPC
		require.NotEmpty(t, rpcEndpoint)

		client, err := rpc.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		desiredSemver, err := SuperchainSemver[superchain.Superchain].VersionFor("ProtocolVersions")
		require.NoError(t, err)
		checkSemverForContract(t, "ProtocolVersions", superchain.Config.ProtocolVersionsAddr, client, desiredSemver)
	}

	for superchainName, superchain := range Superchains {
		t.Run(superchainName, func(t *testing.T) { checkSuperchainTargetSatisfiesSemver(t, superchain) })
	}

}

func TestContractVersions(t *testing.T) {
	isExcluded := map[uint64]bool{
		291:          true,
		424:          true,
		888:          true,
		957:          true,
		997:          true,
		8453:         true,
		34443:        true,
		58008:        true,
		84531:        true,
		84532:        true,
		7777777:      true,
		11155421:     true, // sepolia-dev-0/oplabs-devnet-0
		11763071:     true,
		999999999:    true,
		129831238013: true,
	}

	checkOPChainSatisfiesSemver := func(t *testing.T, chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		require.NotEmpty(t, rpcEndpoint)

		client, err := rpc.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		contractNames := []string{
			"L1CrossDomainMessenger",
			"L1ERC721Bridge",
			"L1StandardBridge",
			"L2OutputOracle",
			"OptimismMintableERC20Factory",
			"OptimismPortal",
			"SystemConfig",
		}

		for _, contractName := range contractNames {

			// ASSUMPTION: we will check the version of the implementation via the declared proxy contract
			proxyContractName := contractName + "Proxy"
			contractAddress, err := Addresses[chain.ChainID].AddressFor(proxyContractName)
			require.NoErrorf(t, err, "%s/%s.%s.version= UNSPECIFIED", chain.Superchain, chain.Name, proxyContractName)

			desiredSemver, err := SuperchainSemver[chain.Superchain].VersionFor(contractName)
			require.NoError(t, err)
			checkSemverForContract(t, proxyContractName, &contractAddress, client, desiredSemver)

			desiredBytecode := []byte{1, 1} // TODO proper ground truth
			checkBytecodeForProxiedContract(t, proxyContractName, &contractAddress, client, desiredBytecode)
		}
	}

	for chainID, chain := range OPChains {
		if isExcluded[chainID] {
			t.Logf("chain %d: EXCLUDED from contract version validation", chainID)
		} else {
			t.Run(chain.Superchain+"/"+chain.Name, func(t *testing.T) { checkOPChainSatisfiesSemver(t, chain) })
		}
	}
}

func checkSemverForContract(t *testing.T, contractName string, contractAddress *Address, client *rpc.Client, desiredSemver string) {
	ethClient := ethclient.NewClient(client)
	actualSemver, err := getVersionWithRetries(context.Background(), common.Address(*contractAddress), ethClient)
	require.NoError(t, err, "Could not get version for %s", contractName)

	require.Condition(t, func() bool { return isSemverAcceptable(desiredSemver, actualSemver) },
		"%s.version=%s (UNACCEPTABLE desired version %s)", contractName, actualSemver, desiredSemver)

	t.Logf("%s.version=%s (acceptable compared to %s)", contractName, actualSemver, desiredSemver)
}

func checkBytecodeForProxiedContract(t *testing.T, contractName string, contractAddress *Address, client *rpc.Client, desiredBytecode []byte) {
	actualBytecode, err := getBytecodeForProxiedContract(context.Background(), common.Address(*contractAddress), client)
	require.NoError(t, err, "Could not get bytecode for %s", contractName)

	require.True(t, bytes.Equal(actualBytecode, desiredBytecode), "unacceptable bytecode for %s, got %s wanted %s", contractName, common.Bytes2Hex(actualBytecode), common.Bytes2Hex(desiredBytecode))

	t.Logf("acceptable bytecode for %s", contractName)
}

// getVersion will get the version of a contract at a given address, if it exposes a version() method.
func getVersion(ctx context.Context, addr common.Address, client *ethclient.Client) (string, error) {
	isemver, err := bindings.NewISemver(addr, client)
	if err != nil {
		return "", fmt.Errorf("%s: %w", addr, err)
	}
	version, err := isemver.Version(&bind.CallOpts{
		Context: ctx,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", addr, err)
	}

	return version, nil
}

// getVersionWithRetries is a wrapper for getVersion
// which retries up to 10 times with exponential backoff.
func getVersionWithRetries(ctx context.Context, addr common.Address, client *ethclient.Client) (string, error) {
	const maxAttempts = 10
	return retry.Do(ctx, maxAttempts, retry.Exponential(), func() (string, error) {
		return getVersion(context.Background(), addr, client)
	})
}

// getBytecodeWithRetries will get the bytecode at a given address, retrying up to 10 times.
func getBytecodeForProxiedContract(ctx context.Context, proxyAddr common.Address, client *rpc.Client) ([]byte, error) {
	const maxAttempts = 1

	implementationAddr, err := getImplementationAddressFromProxy(ctx, proxyAddr, client)
	if err != nil {
		return []byte{}, fmt.Errorf("%s: %w", proxyAddr, err)
	}

	ethClient := ethclient.NewClient(client)
	return retry.Do(ctx, maxAttempts, retry.Exponential(), func() ([]byte, error) {
		return ethClient.CodeAt(ctx, implementationAddr, nil)
	})
}

func getImplementationAddressFromProxy(ctx context.Context, proxyAddr common.Address, client *rpc.Client) (common.Address, error) {
	addr, err := getImplementationAddressFromProxyViaGetter(ctx, proxyAddr, ethclient.NewClient(client))
	if err != nil {
		addr, err = getImplementationAddressFromProxyViaTrace(ctx, proxyAddr, client)
		if err != nil {
			return common.Address{}, fmt.Errorf("could not get implementation address from proxy via trace_call request: %w", err)
		}
	}
	return addr, nil
}

func getImplementationAddressFromProxyViaGetter(ctx context.Context, proxyAddr common.Address, client *ethclient.Client) (common.Address, error) {
	const maxAttempts = 1
	result, err := retry.Do(ctx, maxAttempts, retry.Exponential(), func() ([]byte, error) {
		return client.CallContract(context.Background(), ethereum.CallMsg{
			To:   &proxyAddr,
			From: common.Address{0},             // This should avoid the call being proxied
			Data: common.FromHex("0x5c60da1b")}, // this is the function selector for "Implementation()"
			nil)
	})
	if err != nil {
		return common.Address{}, err
	}
	return common.BytesToAddress(result), nil
}

func getImplementationAddressFromProxyViaTrace(ctx context.Context, proxyAddr common.Address, client *rpc.Client) (common.Address, error) {

	args := map[string]interface{}{
		"to":   proxyAddr.Hex(),
		"data": "0x54fd4d50",
	}

	type Trace struct {
		Action struct {
			CallType string `json:"callType"`
			From     string `json:"from"`
			Gas      string `json:"gas"`
			Input    string `json:"input"`
			To       string `json:"to"`
			Value    string `json:"value"`
		} `json:"action"`
		Result struct {
			GasUsed string `json:"gasUsed"`
			Output  string `json:"output"`
		} `json:"result"`
		Subtraces    int    `json:"subtraces"`
		TraceAddress []int  `json:"traceAddress"`
		Type         string `json:"type"`
	}

	type TraceResult struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			Output    string      `json:"output"`
			StateDiff interface{} `json:"stateDiff"` // Use interface{} type for null
			Trace     []Trace     `json:"trace"`
			VmTrace   interface{} `json:"vmTrace"` // Use interface{} type for null
		} `json:"result"`
	}

	var result TraceResult

	err := client.CallContext(ctx, &result, "trace_call", args, []string{"trace"})

	if err != nil {
		return common.Address{}, err
	}

	for _, trace := range result.Result.Trace {

		if trace.Action.CallType == "delegatecall" {
			return common.HexToAddress(trace.Action.To), nil
		}

	}
	return common.Address{}, fmt.Errorf("could not infer implementation address")
}
