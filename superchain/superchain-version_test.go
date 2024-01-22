package superchain_test

import (
	"context"
	"fmt"
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
func TestContractVersionsCheck(t *testing.T) {
	if err := superchain.SuperchainSemver.SanityCheck(); err != nil {
		t.Fatal(err)
	}

	desiredSemver := superchain.SuperchainSemver.OptimismPortal
	// TODO complete all members of SuperchainSemver
	// TODO resolve missing contracts
	// TODO resolve extraneous contracts
	for _, chain := range superchain.OPChains {
		client, err := ethclient.Dial(superchain.Superchains[chain.Superchain].Config.L1.PublicRPC)
		checkErr(t, err)
		actualSemver, err := getVersion(context.Background(), common.Address(superchain.Addresses[chain.ChainID].OptimismPortalProxy), client) // TODO resolve difference in names
		checkErr(t, err)
		if desiredSemver != actualSemver {
			t.Fatalf("OptimismPortalProxy should have version %v but has version %v (%s on %s)", desiredSemver, actualSemver, chain.Name, chain.Superchain)
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
