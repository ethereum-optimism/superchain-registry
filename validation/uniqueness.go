package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/stretchr/testify/require"
)

type uniqueProperties struct {
	Name      string
	ShortName string
	RPC       []string
}

type chainInfo struct {
	localChainIds   sync.Map
	localChainNames sync.Map
}

var (
	ErrChainIdNotListed        = errors.New("chain ID is not listed at chainid.network")
	ErrLocalChainNameMismatch  = errors.New("local chain name does not match name from chainid.network")
	ErrChainIdDuplicated       = errors.New("chain ID is duplicated")
	ErrChainNameDuplicated     = errors.New("chain name duplicated")
	ErrChainPublicRpcNotListed = errors.New("chain public RPC not listed in chainid.network")
)

func init() {
	var err error
	globalChainIds, err = getGlobalChains()
	if err != nil {
		panic(err)
	}
	localChains = &chainInfo{}
}

func testIsGloballyUnique(t *testing.T, chain *ChainConfig) {
	err := checkIsGloballyUnique(globalChainIds, localChains, chain)
	require.NoError(t, err)
}

func checkIsGloballyUnique(globalIds map[uint]*uniqueProperties, chains *chainInfo, chain *ChainConfig) error {
	props := globalIds[uint(chain.ChainID)]
	if props == nil {
		return ErrChainIdNotListed
	}

	globalChainName := props.Name
	if globalChainName != chain.Name {
		return fmt.Errorf("%w, chainId=%d", ErrLocalChainNameMismatch, chain.ChainID)
	}

	normalizedURL, err := normalizeURL(chain.PublicRPC)
	if err != nil {
		return err
	}

	if !slices.Contains(props.RPC, normalizedURL) {
		return ErrChainPublicRpcNotListed
	}

	if err := chains.AddIfUnique(chain.ChainID, chain.Name); err != nil {
		return err
	}

	return nil
}

func (c *chainInfo) AddIfUnique(chainId uint64, chainName string) error {
	// Check if chainId is unique
	if _, exists := c.localChainIds.LoadOrStore(chainId, true); exists {
		return fmt.Errorf("%w: %d", ErrChainIdDuplicated, chainId)
	}

	// Check if chainName is unique
	if _, exists := c.localChainNames.LoadOrStore(chainName, true); exists {
		return fmt.Errorf("%w: %s", ErrChainNameDuplicated, chainName)
	}

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

	res, err := client.Do(req)
	if err != nil {
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

	err = json.Unmarshal(body, &globalChains)
	if err != nil {
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
