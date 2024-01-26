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
)

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

// TestContractVersions will check that
//   - for each chain in superchain.OPChain
//   - for each declared contract "Foo" : version entry in the corresponding superchain's semver.yaml
//   - the chain has a contract FooProxy deployed at the same version
//
// Actual semvers are
// read from the L1 chain RPC provider for the chain in question.
func TestContractVersions(t *testing.T) {

	isExcluded := map[uint64]bool{
		10:           true,
		291:          true,
		420:          true,
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
		11155420:     true,
		11763071:     true,
		999999999:    true,
		129831238013: true,
	}

	isSemverAcceptable := func(desired, actual string) bool {
		return desired == actual
	}

	checkOPChainSatisfiesSemver := func(chain *superchain.ChainConfig) {
		r := reflect.TypeOf(superchain.SuperchainSemver[chain.Superchain])
		semverFields := reflect.VisibleFields(r)
		for _, field := range semverFields {

			desiredSemver := reflect.Indirect(reflect.ValueOf(superchain.SuperchainSemver[chain.Superchain])).FieldByName(field.Name).String()
			rpcEndpoint := superchain.Superchains[chain.Superchain].Config.L1.PublicRPC
			client, err := ethclient.Dial(rpcEndpoint)
			checkErr(t, err)
			r := reflect.ValueOf(superchain.Addresses[chain.ChainID])
			proxyContractName := field.Name + "Proxy"
			// ASSUMPTION: we will check the version of the implementation via the declared proxy contract
			contractAddressValue := reflect.Indirect(r).FieldByName(proxyContractName)
			if contractAddressValue == (reflect.Value{}) {
				t.Errorf("%10s:%-20s:%-30s has version UNSPECIFIED (desired version %s)", chain.Superchain, chain.Name, proxyContractName, desiredSemver)
				continue
			}
			actualSemver, err := getVersionWithRetries(context.Background(), common.Address(contractAddressValue.Bytes()), client)
			if err != nil {
				t.Errorf("RPC endpoint %s: %s", rpcEndpoint, err)
			}
			if isSemverAcceptable(desiredSemver, actualSemver) {
				t.Logf("%s/%s.%s.version=%s (acceptable compared to %s)", chain.Superchain, chain.Name, proxyContractName, actualSemver, desiredSemver)
			} else {
				t.Errorf("%s/%s.%s.version=%s (UNACCEPTABLE desired version %s)", chain.Superchain, chain.Name, proxyContractName, actualSemver, desiredSemver)
			}

		}
	}

	for chainID, chain := range superchain.OPChains {
		if !isExcluded[chainID] {
			checkOPChainSatisfiesSemver(chain)
		}
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
