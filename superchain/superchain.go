package superchain

import (
	"compress/gzip"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed configs
var superchainFS embed.FS

//go:embed extra/addresses extra/bytecodes extra/genesis extra/genesis-system-configs
var extraFS embed.FS

type BlockID struct {
	Hash   Hash   `yaml:"hash"`
	Number uint64 `yaml:"number"`
}

type ChainGenesis struct {
	L1        BlockID   `yaml:"l1"`
	L2        BlockID   `yaml:"l2"`
	L2Time    uint64    `yaml:"l2_time"`
	ExtraData *HexBytes `yaml:"extra_data,omitempty"`
}

type ChainConfig struct {
	Name         string `yaml:"name"`
	ChainID      uint64 `yaml:"chain_id"`
	PublicRPC    string `yaml:"public_rpc"`
	SequencerRPC string `yaml:"sequencer_rpc"`
	Explorer     string `yaml:"explorer"`

	SystemConfigAddr Address `yaml:"system_config_addr"`
	BatchInboxAddr   Address `yaml:"batch_inbox_addr"`

	Genesis ChainGenesis `yaml:"genesis"`

	// Superchain is a simple string to identify the superchain.
	// This is implied by directory structure, and not encoded in the config file itself.
	Superchain string `yaml:"-"`
	// Chain is a simple string to identify the chain, within its superchain context.
	// This matches the resource filename, it is not encoded in the config file itself.
	Chain string `yaml:"-"`
}

type AddressList struct {
	AddressManager                    Address `json:"AddressManager"`
	L1CrossDomainMessengerProxy       Address `json:"L1CrossDomainMessengerProxy"`
	L1ERC721BridgeProxy               Address `json:"L1ERC721BridgeProxy"`
	L1StandardBridgeProxy             Address `json:"L1StandardBridgeProxy"`
	L2OutputOracleProxy               Address `json:"L2OutputOracleProxy"`
	OptimismMintableERC20FactoryProxy Address `json:"OptimismMintableERC20FactoryProxy"`
	OptimismPortalProxy               Address `json:"OptimismPortalProxy"`
	ProxyAdmin                        Address `json:"ProxyAdmin"`
}

type GenesisSystemConfig struct {
	BatcherAddr Address `json:"batcherAddr"`
	Overhead    Hash    `json:"overhead"`
	Scalar      Hash    `json:"scalar"`
	GasLimit    uint64  `json:"gasLimit"`
}

type GenesisAccount struct {
	CodeHash Hash          `json:"codeHash,omitempty"` // code hash only, to reduce overhead of duplicate bytecode
	Storage  map[Hash]Hash `json:"storage,omitempty"`
	Balance  *HexBig       `json:"balance,omitempty"`
	Nonce    uint64        `json:"nonce,omitempty"`
}

type Genesis struct {
	// Block properties
	Nonce      uint64  `json:"nonce"`
	Timestamp  uint64  `json:"timestamp"`
	ExtraData  []byte  `json:"extraData"`
	GasLimit   uint64  `json:"gasLimit"`
	Difficulty *HexBig `json:"difficulty"`
	Mixhash    Hash    `json:"mixHash"`
	Coinbase   Address `json:"coinbase"`
	Number     uint64  `json:"number"`
	GasUsed    uint64  `json:"gasUsed"`
	ParentHash Hash    `json:"parentHash"`
	BaseFee    *HexBig `json:"baseFeePerGas"`
	// State data
	Alloc map[Address]GenesisAccount `json:"alloc"`
	// StateHash substitutes for a full embedded state allocation,
	// for instantiating states with the genesis block only, to be state-synced before operation.
	// Archive nodes should use a full external genesis.json or datadir.
	StateHash *Hash `json:"stateHash,omitempty"`
	// The chain-config is not included. This is derived from the chain and superchain definition instead.
}

type SuperchainL1Info struct {
	ChainID   uint64 `yaml:"chain_id"`
	PublicRPC string `yaml:"public_rpc"`
	Explorer  string `yaml:"explorer"`
}

type SuperchainConfig struct {
	Name string           `yaml:"name"`
	L1   SuperchainL1Info `yaml:"l1"`

	// TODO: not available yet
	//ProtocolVersionAddr Address `yaml:"protocol_version_addr"`
}

type Superchain struct {
	Config SuperchainConfig

	// Chains that are part of this superchain
	ChainIDs []uint64

	// Superchain identifier, without capitalization or display changes.
	Superchain string
}

var Superchains = map[string]*Superchain{}

var OPChains = map[uint64]*ChainConfig{}

var Addresses = map[uint64]*AddressList{}

var GenesisSystemConfigs = map[uint64]*GenesisSystemConfig{}

func init() {
	superchainTargets, err := superchainFS.ReadDir("configs")
	if err != nil {
		panic(fmt.Errorf("failed to read superchain dir: %w", err))
	}
	// iterate over superchain-target entries
	for _, s := range superchainTargets {
		if !s.IsDir() {
			continue // ignore files, e.g. a readme
		}
		// Load superchain-target config
		superchainConfigData, err := superchainFS.ReadFile(path.Join("configs", s.Name(), "superchain.yaml"))
		if err != nil {
			panic(fmt.Errorf("failed to read superchain config: %w", err))
		}
		var superchainEntry Superchain
		if err := yaml.Unmarshal(superchainConfigData, &superchainEntry.Config); err != nil {
			panic(fmt.Errorf("failed to decode superchain config: %w", err))
		}
		superchainEntry.Superchain = s.Name()

		// iterate over the chains of this superchain-target
		chainEntries, err := superchainFS.ReadDir(path.Join("configs", s.Name()))
		if err != nil {
			panic(fmt.Errorf("failed to read superchain dir: %w", err))
		}
		for _, c := range chainEntries {
			if c.IsDir() || !strings.HasSuffix(c.Name(), ".yaml") {
				continue // ignore files. Chains must be a directory of configs.
			}
			if c.Name() == "superchain.yaml" {
				continue // already processed
			}
			// load chain config
			chainConfigData, err := superchainFS.ReadFile(path.Join("configs", s.Name(), c.Name()))
			if err != nil {
				panic(fmt.Errorf("failed to read superchain config %s/%s: %w", s.Name(), c.Name(), err))
			}
			var chainConfig ChainConfig
			if err := yaml.Unmarshal(chainConfigData, &chainConfig); err != nil {
				panic(fmt.Errorf("failed to decode chain config %s/%s: %w", s.Name(), c.Name(), err))
			}
			chainConfig.Chain = strings.TrimSuffix(c.Name(), ".yaml")

			jsonName := chainConfig.Chain + ".json"
			addressesData, err := extraFS.ReadFile(path.Join("extra", "addresses", s.Name(), jsonName))
			if err != nil {
				panic(fmt.Errorf("failed to read addresses data of chain %s/%s: %w", s.Name(), jsonName, err))
			}
			var addrs AddressList
			if err := json.Unmarshal(addressesData, &addrs); err != nil {
				panic(fmt.Errorf("failed to decode addresses %s/%s: %w", s.Name(), jsonName, err))
			}

			genesisSysCfgData, err := extraFS.ReadFile(path.Join("extra", "genesis-system-configs", s.Name(), jsonName))
			if err != nil {
				panic(fmt.Errorf("failed to read genesis system config data of chain %s/%s: %w", s.Name(), jsonName, err))
			}
			var genesisSysCfg GenesisSystemConfig
			if err := json.Unmarshal(genesisSysCfgData, &genesisSysCfg); err != nil {
				panic(fmt.Errorf("failed to decode genesis system config %s/%s: %w", s.Name(), jsonName, err))
			}

			chainConfig.Superchain = s.Name()
			if other, ok := OPChains[chainConfig.ChainID]; ok {
				panic(fmt.Errorf("found chain config %q in superchain target %q with chain ID %d "+
					"conflicts with chain %q in superchain %q and chain ID %d",
					chainConfig.Name, chainConfig.Superchain, chainConfig.ChainID,
					other.Name, other.Superchain, other.ChainID))
			}
			superchainEntry.ChainIDs = append(superchainEntry.ChainIDs, chainConfig.ChainID)
			OPChains[chainConfig.ChainID] = &chainConfig
			Addresses[chainConfig.ChainID] = &addrs
			GenesisSystemConfigs[chainConfig.ChainID] = &genesisSysCfg
		}
		Superchains[superchainEntry.Superchain] = &superchainEntry
	}
}

func LoadGenesis(chainID uint64) (*Genesis, error) {
	ch, ok := OPChains[chainID]
	if !ok {
		return nil, fmt.Errorf("unknown chain %d", chainID)
	}
	f, err := extraFS.Open(path.Join("extra", "genesis", ch.Superchain, ch.Chain+".json.gz"))
	if err != nil {
		return nil, fmt.Errorf("failed to open chain genesis definition of %d: %w", chainID, err)
	}
	defer f.Close()
	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip reader of genesis data of %d: %w", chainID, err)
	}
	defer r.Close()
	var out Genesis
	if err := json.NewDecoder(r).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode genesis allocation of %d: %w", chainID, err)
	}
	return &out, nil
}

func LoadContractBytecode(codeHash Hash) ([]byte, error) {
	f, err := extraFS.Open(path.Join("extra", "bytecodes", codeHash.String()+".bin.gz"))
	if err != nil {
		return nil, fmt.Errorf("failed to open bytecode %s: %w", codeHash, err)
	}
	defer f.Close()
	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("")
	}
	defer r.Close()
	return io.ReadAll(r)
}
