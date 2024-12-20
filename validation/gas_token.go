package validation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func testGasToken(t *testing.T, chain *superchain.ChainConfig) {
	l1Client, err := ethclient.Dial(superchain.Superchains[chain.Superchain].Config.L1.PublicRPC)
	require.NoError(t, err, "Failed to connect to L1 EthClient at RPC url %s", superchain.Superchains[chain.Superchain].Config.L1.PublicRPC)
	defer l1Client.Close()

	l2Client, err := ethclient.Dial(chain.PublicRPC)
	require.NoError(t, err, "Failed to connect to L2 EthClient at RPC url %s", chain.PublicRPC)
	defer l2Client.Close()

	err = CheckGasToken(chain, l1Client, l2Client)
	require.NoError(t, err)
}

func CheckGasToken(chain *superchain.ChainConfig, l1Client EthClient, l2Client EthClient) error {
	weth9PredeployAddress := superchain.MustHexToAddress("0x4200000000000000000000000000000000000006")
	want := "0000000000000000000000000000000000000000000000000000000000000020" + // offset
		"000000000000000000000000000000000000000000000000000000000000000d" + // length
		"5772617070656420457468657200000000000000000000000000000000000000" // "Wrapped Ether" padded to 32 bytes
	gotName, err := getHexString("name()", weth9PredeployAddress, l2Client)
	if err != nil {
		return err
	}
	if want != gotName {
		return fmt.Errorf("predeploy WETH9.name(): want=%s, got=%s", want, gotName)
	}

	l1BlockPredeployAddress := superchain.MustHexToAddress("0x4200000000000000000000000000000000000015")
	isCustomGasToken, err := getBool("isCustomGasToken()", l1BlockPredeployAddress, l2Client)
	if err != nil && !strings.Contains(err.Error(), "execution reverted") {
		// Pre: reverting is acceptable
		return err
	} else {
		// Post: must be set to false
		if isCustomGasToken {
			return fmt.Errorf("L1Block.isCustomGasToken() must return false")
		}
	}

	isCustomGasToken, err = getBool("isCustomGasToken()", superchain.Addresses[chain.ChainID].SystemConfigProxy, l1Client)
	if err != nil && !strings.Contains(err.Error(), "execution reverted") {
		// Pre: reverting is acceptable
		return err
	} else {
		// Post: must be set to false
		if isCustomGasToken {
			return fmt.Errorf("SystemConfigProxy.isCustomGasToken() must return false")
		}
	}
	return nil
}

func getBytes(method string, contractAddress superchain.Address, client EthClient) ([]byte, error) {
	addr := (common.Address(contractAddress))
	callMsg := ethereum.CallMsg{
		To:   &addr,
		Data: crypto.Keccak256([]byte(method))[:4],
	}

	callContract := func(msg ethereum.CallMsg) ([]byte, error) {
		return client.CallContract(context.Background(), msg, nil)
	}

	return Retry(callContract)(callMsg)
}

func getHexString(method string, contractAddress superchain.Address, client EthClient) (string, error) {
	result, err := getBytes(method, contractAddress, client)
	return common.Bytes2Hex(result), err
}

func getBool(method string, contractAddress superchain.Address, client EthClient) (bool, error) {
	result, err := getBytes(method, contractAddress, client)
	if err != nil {
		return false, err
	}

	switch common.HexToHash(string(result)) {
	case common.Hash{1}:
		return true, nil
	case common.Hash{}:
		return false, nil
	default:
		return false, errors.New("unexpected non-bool return value")
	}
}
