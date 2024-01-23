package superchain_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

// TestContractVersionsCheck will fail if the superchain semver file
// is not read correctly.
func TestSemverFile(t *testing.T) {
	if err := superchain.SuperchainSemver.SanityCheck(); err != nil {
		t.Fatal(err)
	}
}

// TestContractVersionsCheck will check that
//   - for each declared contract "Foo" in superchain.SuperchainSemver
//   - for each chain in superchain.OPChains
//
// there is a contract address declared for "FooProxy" which has an
// actual semver matching the declared semver. Actual semvers are
// read from the L1 chain RPC provider for the chain in question.
func TestContractVersionsCheck(t *testing.T) {

	checkAllOPChainsSatisfySemver := func(contractName string) {
		desiredSemver := reflect.ValueOf(superchain.SuperchainSemver).FieldByName(contractName).String()

		// ASSUMPTION: we will check the version of the implementation via the declared proxy contract
		proxyContractName := contractName + "Proxy"

		for _, chain := range superchain.OPChains {
			rpcEndpoint := superchain.Superchains[chain.Superchain].Config.L1.PublicRPC
			t.Logf("Dialing %s...", rpcEndpoint)
			client, err := ethclient.Dial(rpcEndpoint)
			checkErr(t, err)
			r := reflect.ValueOf(superchain.Addresses[chain.ChainID])
			contractAddressValue := reflect.Indirect(r).FieldByName(proxyContractName)
			if contractAddressValue == (reflect.Value{}) {
				t.Fatalf("Semver for %s not specified for chain %s on %s", proxyContractName, chain.Name, chain.Superchain)
			}
			actualSemver, err := getVersion(context.Background(), common.Address(contractAddressValue.Bytes()), client)
			checkErr(t, err)
			if desiredSemver != actualSemver {
				t.Fatalf("%v should have version %v but has version %v (%s on %s)", contractName, desiredSemver, actualSemver, chain.Name, chain.Superchain)
			}
		}
	}

	for i := 0; i < reflect.ValueOf(superchain.SuperchainSemver).NumField(); i++ {
		contractName := reflect.ValueOf(superchain.SuperchainSemver).Field(i).String()
		checkAllOPChainsSatisfySemver(contractName)
	}

}

// getVersion will get the version of a contract at a given address, if it exposes a version() method.
func getVersion(ctx context.Context, addr common.Address, client *ethclient.Client) (string, error) {
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &addr,
		Data: common.FromHex("0x54fd4d50")}, // this is the function selector for "version"
		nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", addr, err)
	}
	var String, _ = abi.NewType("string", "string", nil)

	s, err := abi.Arguments{abi.Argument{Type: String}}.Unpack(result)
	if err != nil {
		panic(err)
	}

	return s[0].(string), nil
}
