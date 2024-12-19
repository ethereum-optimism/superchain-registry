package validation

import (
	"fmt"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testSuperchainConfig(t *testing.T, chain *ChainConfig) {
	rpcEndpoint := Superchains[chain.Superchain].Config.L1.PublicRPC
	require.NotEmpty(t, rpcEndpoint, "no rpc specified")

	l1Client, err := ethclient.Dial(rpcEndpoint)
	require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)
	defer l1Client.Close()

	err = checkSuperchainConfig(chain, l1Client)
	require.NoError(t, err)
}

func checkSuperchainConfig(chain *ChainConfig, l1Client *ethclient.Client) error {
	expected := Superchains[chain.Superchain].Config.SuperchainConfigAddr
	if expected == nil {
		return fmt.Errorf("Superchain does not declare a superchain_config_addr")
	}

	opcm := Superchains[chain.Superchain].Config.OPContractsManagerProxyAddr
	if opcm == nil {
		return fmt.Errorf("Superchain does not declare a op_contracts_manager_proxy_addr")
	}

	contracts := []Address{
		Addresses[chain.ChainID].OptimismPortalProxy,
		Addresses[chain.ChainID].AnchorStateRegistryProxy,
		Addresses[chain.ChainID].L1CrossDomainMessengerProxy,
		Addresses[chain.ChainID].L1ERC721BridgeProxy,
		Addresses[chain.ChainID].L1StandardBridgeProxy,
		Addresses[chain.ChainID].DelayedWETHProxy,
	}
	if err := verifySuperchainConfigOnchain(chain.ChainID, l1Client, contracts, *expected); err != nil {
		return err
	}
	return nil
}

func verifySuperchainConfigOnchain(chainID uint64, l1Client EthClient, targetContracts []Address, expected Address) error {
	for _, target := range targetContracts {
		var got Address
		var err error
		if target == Addresses[chainID].DelayedWETHProxy {
			// DelayedWETHProxy uses a different method name, so it is broken out here
			got, err = getAddressFromChain("config()", target, l1Client)
		} else {
			got, err = getAddressFromChain("superchainConfig()", target, l1Client)
		}

		if err != nil {
			return err
		}

		if expected != got {
			return fmt.Errorf("incorrect superchainConfig() address: got %s, wanted %s (queried %s)", got, expected, target)
		}
	}
	return nil
}
