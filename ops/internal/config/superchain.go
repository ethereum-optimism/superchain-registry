package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

type Superchain string

const (
	MainnetSuperchain     Superchain = "mainnet"
	SepoliaSuperchain     Superchain = "sepolia"
	SepoliaDev0Superchain Superchain = "sepolia-dev-0"
)

func ParseSuperchain(in string) (Superchain, error) {
	return Superchain(in), nil
}

func MustParseSuperchain(in string) Superchain {
	sup, err := ParseSuperchain(in)
	if err != nil {
		panic(err)
	}
	return sup
}

// FindValidL1URL finds a valid l1-rpc-url for a given superchain by finding matching l1 chainId
func FindValidL1URL(ctx context.Context, lgr log.Logger, urls []string, superchainId uint64) (string, error) {
	lgr.Info("searching for valid l1-rpc-url", "superchainId", superchainId)
	for i, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}

		if err := validateL1ChainID(ctx, url, superchainId); err != nil {
			lgr.Warn("l1-rpc-url has mismatched l1 chainId", "urlIndex", i)
			continue
		}

		lgr.Info("l1-rpc-url has matching l1 chainId", "urlIndex", i)
		return url, nil
	}
	return "", fmt.Errorf("no valid L1 RPC URL found for superchain %d", superchainId)
}

// validateL1ChainID checks if the l1RpcUrl has the expected chain ID for the superchain
func validateL1ChainID(ctx context.Context, l1RpcUrl string, superchainId uint64) error {
	chainID, err := getL1ChainId(ctx, l1RpcUrl)
	if err != nil {
		return fmt.Errorf("failed to get chainId from l1RpcUrl: %w", err)
	}

	if chainID != superchainId {
		return fmt.Errorf("l1RpcUrl chainId mismatch: got %d, expected %d", chainID, superchainId)
	}

	return nil
}

// getL1ChainId connects to an Ethereum RPC endpoint and retrieves its chain ID
func getL1ChainId(ctx context.Context, rpcURL string) (uint64, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to L1 RPC: %w", err)
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return chainID.Uint64(), nil
}

func (s *Superchain) UnmarshalText(text []byte) error {
	sup, err := ParseSuperchain(string(text))
	if err != nil {
		return err
	}
	*s = sup
	return nil
}

type SuperchainDefinition struct {
	Name                   string              `toml:"name"`
	ProtocolVersionsAddr   *ChecksummedAddress `toml:"protocol_versions_addr"`
	SuperchainConfigAddr   *ChecksummedAddress `toml:"superchain_config_addr"`
	OPContractsManagerAddr *ChecksummedAddress `toml:"op_contracts_manager_addr"`
	Hardforks              Hardforks           `toml:"hardforks"`
	L1                     SuperchainL1        `toml:"l1"`
}

type SuperchainL1 struct {
	ChainID   uint64 `toml:"chain_id"`
	PublicRPC string `toml:"public_rpc"`
	Explorer  string `toml:"explorer"`
}
