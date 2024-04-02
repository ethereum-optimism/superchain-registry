package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
)

func TestGloballyUniqueChainId(t *testing.T) {
	globalChainIds, err := getGlobalChainIds()
	if err != nil {
		t.Fatal(err)
	}

	for _, chain := range OPChains {
		if globalChainIds[uint(chain.ChainID)] != "" &&
			globalChainIds[uint(chain.ChainID)] != chain.Name {
			t.Errorf("Chain %s with ID %d exists in global chain list with name %s",
				chain.Name, chain.ChainID, globalChainIds[uint(chain.ChainID)])
		}
	}

}

func getGlobalChainIds() (map[uint]string, error) {
	chainListUrl := "https://chainid.network/chains_mini.json"

	client := http.Client{}

	req, err := http.NewRequest(http.MethodGet, chainListUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "optimism-superchain-registry-validation")

	res, getErr := client.Do(req)
	if getErr != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	type entry struct {
		ChainId uint   `json:"chainId"`
		Name    string `json:"name"`
	}

	globalChains := make([]entry, 1000)

	jsonErr := json.Unmarshal(body, &globalChains)
	if jsonErr != nil {
		return nil, err
	}

	globalChainIds := make(map[uint]string)
	for _, chain := range globalChains {
		if globalChainIds[chain.ChainId] != "" {
			return nil, fmt.Errorf("Chains listed at %s have duplicate chain Id %d",
				chainListUrl, chain.ChainId)
		}
		globalChainIds[chain.ChainId] = chain.Name
	}
	return globalChainIds, nil
}
