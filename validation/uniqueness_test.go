package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type uniqueProperties struct {
	Name      string
	ShortName string
}

type chainIDSet map[uint64]bool

func (s chainIDSet) AddIfUnique(id uint64) error {
	if s[id] {
		return fmt.Errorf("Chain with ID %d is duplicated", id)
	}
	s[id] = true
	return nil
}

type chainNameSet map[string]bool

func (s chainNameSet) AddIfUnique(name string) error {
	if s[name] {
		return fmt.Errorf("Chain with name %s is duplicated", name)
	}
	s[name] = true
	return nil
}

type chainShortNameSet map[string]bool

func (s chainShortNameSet) AddIfUnique(name string) error {
	if s[name] {
		return fmt.Errorf("Chain with short name %s is duplicated", name)
	}
	s[name] = true
	return nil
}

func getGlobalChains() (map[uint]*uniqueProperties, error) {
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
		ChainId   uint   `json:"chainId"`
		Name      string `json:"name"`
		ShortName string `json:"shortName"`
	}

	globalChains := make([]entry, 1000)

	jsonErr := json.Unmarshal(body, &globalChains)
	if jsonErr != nil {
		return nil, err
	}

	globalChainIds := make(map[uint]*uniqueProperties)
	for _, chain := range globalChains {
		if globalChainIds[chain.ChainId] != nil {
			return nil, fmt.Errorf("Chains listed at %s have duplicate chain Id %d",
				chainListUrl, chain.ChainId)
		}
		globalChainIds[chain.ChainId] = &uniqueProperties{chain.Name, chain.ShortName}
	}
	return globalChainIds, nil
}

func TestChainsAreGloballyUnique(t *testing.T) {
	globalChainIds, err := getGlobalChains()
	if err != nil {
		t.Fatal(err)
	}
	localChainIds := make(chainIDSet)
	localChainNames := make(chainNameSet)
	localChainShortNames := make(chainShortNameSet)

	isExcluded := map[uint64]bool{
		90001: true, // sepolia/race, known chainId collision
	}

	for _, chain := range OPChains {
		t.Run(perChainTestName(chain), func(t *testing.T) {
			if isExcluded[chain.ChainID] {
				t.Skip("excluded from global chain id check")
			}
			SkipCheckIfDevnet(t, *chain)
			require.NotNil(t, globalChainIds[uint(chain.ChainID)], "chain ID is not listed at chainid.network")
			globalChainName := globalChainIds[uint(chain.ChainID)].Name
			globalShortName := globalChainIds[uint(chain.ChainID)].ShortName
			assert.Equal(t, globalChainName, chain.Name,
				"Local chain name for %d does not match name from chainid.network", chain.ChainID)
			assert.Equal(t, chain.ShortName, globalShortName,
				"Local short chain name for %d does not match name from chainid.network", chain.ChainID)

			assert.NoError(t, localChainIds.AddIfUnique(chain.ChainID))
			assert.NoError(t, localChainNames.AddIfUnique(chain.Name))
			assert.NoError(t, localChainShortNames.AddIfUnique(chain.ShortName))
		})
	}
}
