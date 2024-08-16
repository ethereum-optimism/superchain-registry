package genesis

import (
	"github.com/ethereum-optimism/superchain-registry/superchain"
)

// Define a struct to represent the structure of the JSON data
type DeployedBytecode struct {
	Object              string                          `json:"object"`
	ImmutableReferences map[string][]ImmutableReference `json:"immutableReferences"`
}

type ImmutableReference struct {
	Start  int `json:"start"`
	Length int `json:"length"`
}

type ContractData struct {
	DeployedBytecode DeployedBytecode `json:"deployedBytecode"`
}

type GenesisAccountLite struct {
	Code    string             `json:"code,omitempty"`
	Storage map[string]string  `json:"storage,omitempty"`
	Balance *superchain.HexBig `json:"balance,omitempty"`
	Nonce   uint64             `json:"nonce,omitempty"`
}

type GenesisLite struct {
	// State data
	Alloc map[string]GenesisAccountLite `json:"alloc"`
}
