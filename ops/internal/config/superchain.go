package config

type Superchain struct {
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
