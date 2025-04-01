package config

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Superchain string

const (
	MainnetSuperchain     Superchain = "mainnet"
	SepoliaSuperchain     Superchain = "sepolia"
	SepoliaDev0Superchain Superchain = "sepolia-dev-0"
)

var SuperchainChainIds = map[Superchain]uint64{
	MainnetSuperchain:     1,
	SepoliaSuperchain:     11155111,
	SepoliaDev0Superchain: 11155111,
}

func ParseSuperchain(in string) (Superchain, error) {
	switch Superchain(in) {
	case MainnetSuperchain, SepoliaSuperchain, SepoliaDev0Superchain:
		return Superchain(in), nil
	default:
		return "", fmt.Errorf("unknown superchain: '%s'", in)
	}
}

func MustParseSuperchain(in string) Superchain {
	sup, err := ParseSuperchain(in)
	if err != nil {
		panic(err)
	}
	return sup
}

// ValidateL1ChainID checks if the L1 RPC URL has the expected chain ID for the superchain
func ValidateL1ChainID(l1RPCURL string, superchain Superchain) error {
	// Determine expected chain ID based on superchain
	expectedChainID, ok := SuperchainChainIds[superchain]
	if !ok {
		return fmt.Errorf("unknown superchain: %s", superchain)
	}

	// Get the actual chain ID
	chainID, err := GetL1ChainID(l1RPCURL)
	if err != nil {
		return fmt.Errorf("failed to get chain ID from L1 RPC at %s: %w", l1RPCURL, err)
	}

	// Compare with expected chain ID
	if chainID != expectedChainID {
		return fmt.Errorf("L1 RPC chain ID mismatch: got %d, expected %d for superchain %s",
			chainID, expectedChainID, superchain)
	}

	return nil
}

// GetL1ChainID connects to an Ethereum RPC endpoint and retrieves its chain ID
func GetL1ChainID(rpcURL string) (uint64, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to L1 RPC: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return chainID.Uint64(), nil
}

func GetSuperchainChainId(superchain string) (uint64, error) {
	validatedSuperchain, err := ParseSuperchain(superchain)
	if err != nil {
		return 0, fmt.Errorf("error parsing superchain: %w", err)
	}

	switch validatedSuperchain {
	case MainnetSuperchain:
		return 1, nil
	case SepoliaSuperchain, SepoliaDev0Superchain:
		return 11155111, nil
	default:
		return 0, fmt.Errorf("unknown superchain: '%s'", superchain)
	}
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
	Name                   string
	ProtocolVersionsAddr   *ChecksummedAddress `toml:"protocol_versions_addr"`
	SuperchainConfigAddr   *ChecksummedAddress `toml:"superchain_config_addr"`
	OPContractsManagerAddr *ChecksummedAddress `toml:"op_contracts_manager_addr"`
	Hardforks              Hardforks           `toml:"hardforks"`
	L1                     SuperchainL1        `toml:"l1"`
}

type SuperchainL1 struct {
	ChainID   uint64
	PublicRPC string `toml:"public_rpc"`
	Explorer  string `toml:"explorer"`
}
