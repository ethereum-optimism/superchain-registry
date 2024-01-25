package validation

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/mod/semver"
)

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

const SOURCE_OF_TRUTH_CHAINID = 10
const SOURCE_OF_TRUTH_SUPERCHAIN = "mainnet"

// TestContractVersions will check that
//   - for each declared contract "FooProxy" in superchain.Addresses[SOURCE_OF_TRUTH_CHAINID]
//   - for each chain in superchain.OPChain
//
// there is a contract address declared for "FooProxy" which has an
// actual semver matching the desired semver. Actual semvers are
// read from the L1 chain RPC provider for the chain in question.
// Desired semvers are read from the SOURCE_OF_TRUTH_SUPERCHAIN.
func TestContractVersions(t *testing.T) {

	isSemverAcceptable := func(desired, actual string) bool {
		return semver.Compare(desired, actual) <= 0
	}

	semverFields := reflect.VisibleFields(reflect.TypeOf(superchain.SuperchainSemver))

	desiredSemver := map[string]string{}
	for _, field := range semverFields {
		if field.Name == "SystemConfig" {
			continue // TODO this is specified in semver.yaml but not in an address list.
		}
		proxyContractName := field.Name + "Proxy"
		r := reflect.ValueOf(superchain.Addresses[SOURCE_OF_TRUTH_CHAINID])
		proxyContractAddressValue := reflect.Indirect(r).FieldByName(proxyContractName)
		client, err := ethclient.Dial(superchain.Superchains[SOURCE_OF_TRUTH_SUPERCHAIN].Config.L1.PublicRPC)
		checkErr(t, err)
		actualSemver, err := getVersionWithRetries(context.Background(), common.Address(proxyContractAddressValue.Bytes()), client)
		checkErr(t, err)
		desiredSemver[field.Name] = actualSemver
	}

	checkOPChainSatisfiesSemver := func(chain *superchain.ChainConfig) {

		for _, field := range semverFields {
			if field.Name == "SystemConfig" {
				continue // TODO this is specified in semver.yaml but not in an address list.
			}
			proxyContractName := field.Name + "Proxy"
			// ASSUMPTION: we will check the version of the implementation via the declared proxy contract
			desiredSemver := desiredSemver[field.Name]
			rpcEndpoint := superchain.Superchains[chain.Superchain].Config.L1.PublicRPC
			client, err := ethclient.Dial(rpcEndpoint)
			checkErr(t, err)
			r := reflect.ValueOf(superchain.Addresses[chain.ChainID])
			contractAddressValue := reflect.Indirect(r).FieldByName(proxyContractName)
			if contractAddressValue == (reflect.Value{}) {
				t.Errorf("%10s:%-20s:%-30s has version UNSPECIFIED (desired version %s)", chain.Superchain, chain.Name, proxyContractName, desiredSemver)
				continue
			}
			actualSemver, err := getVersionWithRetries(context.Background(), common.Address(contractAddressValue.Bytes()), client)
			if err != nil {
				t.Errorf("RPC endpoint %s: %s", rpcEndpoint, err)
			}
			if !isSemverAcceptable(desiredSemver, actualSemver) {
				t.Errorf("%10s:%-20s:%-30s has version %s (desired version %s)", chain.Superchain, chain.Name, proxyContractName, actualSemver, desiredSemver)
			}
		}
	}

	for _, chain := range superchain.OPChains {
		checkOPChainSatisfiesSemver(chain)
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

// getVersionWithRetries is a wrapper for getVersion
// which, on error, will wait 5 seconds an retry up to 3 times.
func getVersionWithRetries(ctx context.Context, addr common.Address, client *ethclient.Client) (string, error) {
	numRetries := 3
	retriesRemaining := numRetries
	s, err := getVersion(ctx, addr, client)
	for err != nil && retriesRemaining > 0 {
		<-time.After(5 * time.Second)
		retriesRemaining--
		s, err = getVersion(ctx, addr, client)
	}
	if err != nil {
		return "", fmt.Errorf("getVersion request failed after %d attempts: %w", numRetries, err)
	}
	return s, nil
}
