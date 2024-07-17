package standard

import "github.com/ethereum-optimism/superchain-registry/superchain"

// Config.Params is keyed by superchain target, e.g. "mainnet" or "sepolia" or "sepolia-dev-0"
type ConfigType struct {
	GenesisAlloc  map[superchain.Address]superchain.GenesisAccount `json:"alloc"`
	Params        map[string]*Params
	Roles         *Roles
	MultisigRoles map[string]*MultisigRoles
}

var Config = ConfigType{}
