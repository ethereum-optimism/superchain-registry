package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type uniqueProperties struct {
	Name      string
	ShortName string
	RPC       []string
}

type (
	chainIDSet   map[uint64]bool
	chainNameSet map[string]bool
)

type chainInfo struct {
	localChainIds   chainIDSet
	localChainNames chainNameSet
	chainIDMutex    sync.RWMutex
	chainNameMutex  sync.RWMutex
}

func (c *chainInfo) AddIfUnique(chainId uint64, chainName string) error {
	// Check if chainId is unique
	c.chainIDMutex.RLock()
	if c.localChainIds[chainId] {
		c.chainIDMutex.RUnlock()
		return fmt.Errorf("Chain with ID %d is duplicated", chainId)
	}
	c.chainIDMutex.RUnlock()

	// Check if chainName is unique
	c.chainNameMutex.RLock()
	if c.localChainIds[chainId] {
		c.chainNameMutex.RUnlock()
		return fmt.Errorf("Chain with name %s is duplicated", chainName)
	}
	c.chainNameMutex.RUnlock()

	c.chainNameMutex.Lock()
	c.localChainNames[chainName] = true
	c.chainNameMutex.Unlock()

	return nil
}

func getGlobalChains() (map[uint]*uniqueProperties, error) {
	// The following URL exposes the list of chains from
	// https://github.com/ethereum-lists/chains, which is
	// the leading repository of EVM chain information.
	// It is where EVM chains go to claim a chainID and to
	// becomes discoverable by wallets and users in general.
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
		ChainId   uint     `json:"chainId"`
		Name      string   `json:"name"`
		ShortName string   `json:"shortName"`
		RPC       []string `json:"rpc"`
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
		for i, url := range chain.RPC {
			normalizedURL, err := normalizeURL(url)
			if err != nil {
				return nil, err
			}
			chain.RPC[i] = normalizedURL
		}
		globalChainIds[chain.ChainId] = &uniqueProperties{chain.Name, chain.ShortName, chain.RPC}
	}
	return globalChainIds, nil
}

var (
	globalChainIds map[uint]*uniqueProperties
	localChains    *chainInfo
)

func init() {
	var err error
	globalChainIds, err = getGlobalChains()
	if err != nil {
		panic(err)
	}
	localChains = &chainInfo{
		localChainIds:   make(chainIDSet),
		localChainNames: make(chainNameSet),
	}
}

func testIsGloballyUnique(t *testing.T, chain *ChainConfig) {
	props := globalChainIds[uint(chain.ChainID)]
	require.NotNil(t, props, "chain ID is not listed at chainid.network")
	globalChainName := props.Name
	assert.Equal(t, globalChainName, chain.Name,
		"Local chain name for %d does not match name from chainid.network", chain.ChainID)
	assert.NoError(t, localChains.AddIfUnique(chain.ChainID, chain.Name))
	normalizedURL, err := normalizeURL(chain.PublicRPC)
	require.NoError(t, err)
	assert.Contains(t, props.RPC, normalizedURL, "Specified RPC not specified in chainid.network")
}

func normalizeURL(rawURL string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Convert scheme and host to lowercase
	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	parsedURL.Host = strings.ToLower(parsedURL.Host)

	// Ensure the path ends with a slash
	if parsedURL.Path == "" {
		parsedURL.Path = "/"
	} else {
		parsedURL.Path = strings.ReplaceAll(parsedURL.Path, "//", "/")
		if !strings.HasSuffix(parsedURL.Path, "/") {
			parsedURL.Path += "/"
		}
	}

	// Remove default port (if any)
	if (parsedURL.Scheme == "http" && parsedURL.Port() == "80") || (parsedURL.Scheme == "https" && parsedURL.Port() == "443") {
		parsedURL.Host = parsedURL.Hostname()
	}

	// Reassemble the URL
	normalizedURL := parsedURL.String()

	return normalizedURL, nil
}

func TestNormalizeURL(t *testing.T) {
	tt := []struct {
		input string
		want  string
	}{
		{"https://rpc.zora.energy/", "https://rpc.zora.energy/"},
		{"https://rpc.zora.energy", "https://rpc.zora.energy/"},
	}

	for _, tc := range tt {
		got, err := normalizeURL(tc.input)
		require.NoError(t, err)
		require.Equal(t, tc.want, got)
	}
}
