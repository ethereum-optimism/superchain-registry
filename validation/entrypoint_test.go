package validation

import (
	"os"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum/go-ethereum/ethclient"
)

var clients = struct {
	L1 map[string]*ethclient.Client
	L2 map[uint64]*ethclient.Client
}{
	L1: make(map[string]*ethclient.Client),
	L2: make(map[uint64]*ethclient.Client),
}

func getRPCClient(rpcUrl string) *ethclient.Client {
	if rpcUrl == "" {
		panic("empty rpcUrl")
	}
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		panic(err)
	}
	return client
}

func TestMain(m *testing.M) {

	for k, v := range superchain.Superchains {
		if v.Config.L1.PublicRPC == "" {
			panic(k + " has no rpc url specified")
		}
		clients.L1[k] = getRPCClient(v.Config.L1.PublicRPC)
	}

	for k, v := range superchain.OPChains {
		if v.PublicRPC == "" {
			continue // it might be a devnet
			// panic("chain with ID " + fmt.Sprint(k) + " has no public rpc url")
		}
		clients.L2[k] = getRPCClient(v.PublicRPC)
	}

	code := m.Run()

	// Cleanup code: Close the client
	for _, v := range clients.L1 {
		v.Close()
	}
	for _, v := range clients.L2 {
		v.Close()
	}

	// Exit with the test code
	os.Exit(code)
}
