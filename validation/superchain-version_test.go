package validation

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	. "github.com/ethereum-optimism/superchain-registry/superchain"

	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TestContractVersions will check that
//   - for each chain in OPChain
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

	checkOPChainSatisfiesSemver := func(chain *ChainConfig) {
		rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC

		if rpcEndpoint == "" {
			t.Errorf("%s has MISSING RPC endpoint", chain.Superchain)
			return
		}

		client, err := ethclient.Dial(rpcEndpoint)

		if err != nil {
			t.Errorf("could not dial rpc endpoint %s: %v", rpcEndpoint, err)
			return
		}

		semverFields := reflect.VisibleFields(reflect.TypeOf(SuperchainSemver[chain.Superchain]))

		for _, field := range semverFields {
			desiredSemver := reflect.Indirect(reflect.ValueOf(SuperchainSemver[chain.Superchain])).FieldByName(field.Name).String()

			// ASSUMPTION: we will check the version of the implementation via the declared proxy contract
			proxyContractName := field.Name + "Proxy"

			contractAddressValue := reflect.Indirect(reflect.ValueOf(Addresses[chain.ChainID])).FieldByName(proxyContractName)
			if contractAddressValue == (reflect.Value{}) {
				t.Errorf("%s/%s.%s.version= UNSPECIFIED (desired version %s)", chain.Superchain, chain.Name, proxyContractName, desiredSemver)
				continue
			}

			actualSemver, err := getVersionWithRetries(context.Background(), common.Address(contractAddressValue.Bytes()), client)
			if err != nil {
				t.Errorf("RPC endpoint %s: %s", rpcEndpoint, err)
				continue
			}

			if isSemverAcceptable(desiredSemver, actualSemver) {
				t.Logf("%s/%s.%s.version=%s (acceptable compared to %s)", chain.Superchain, chain.Name, proxyContractName, actualSemver, desiredSemver)
			} else {
				t.Errorf("%s/%s.%s.version=%s (UNACCEPTABLE desired version %s)", chain.Superchain, chain.Name, proxyContractName, actualSemver, desiredSemver)
			}

		}
	}

	for chainID, chain := range OPChains {
		if !isExcluded[chainID] {
			checkOPChainSatisfiesSemver(chain)
		}
	}
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
	return retry.Do(context.Background(), maxAttempts, retry.Exponential(), func() (string, error) {
		return getVersion(context.Background(), addr, client)
	})
}
