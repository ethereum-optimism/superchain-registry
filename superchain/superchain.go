package superchain

import (
	"compress/gzip"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

//go:embed configs
var superchainFS embed.FS

//go:embed extra/addresses extra/bytecodes extra/genesis extra/genesis-system-configs
var extraFS embed.FS

//go:embed implementations/networks implementations
var implementationsFS embed.FS

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

// ContractImplementations represent a set of contract implementations on a given network.
// The key in the map represents the semantic version of the contract and the value is the
// address that the contract is deployed to.
type ContractImplementations struct {
	L1CrossDomainMessenger       map[string]Address `yaml:"l1_cross_domain_messenger"`
	L1ERC721Bridge               map[string]Address `yaml:"l1_erc721_bridge"`
	L1StandardBridge             map[string]Address `yaml:"l1_standard_bridge"`
	L2OutputOracle               map[string]Address `yaml:"l2_output_oracle"`
	OptimismMintableERC20Factory map[string]Address `yaml:"optimism_mintable_erc20_factory"`
	OptimismPortal               map[string]Address `yaml:"optimism_portal"`
	SystemConfig                 map[string]Address `yaml:"system_config"`
}

// NewContractImplementations returns a new empty ContractImplementations.
// Use this constructor to ensure that none of struct fields are nil.
// The path argument is relative to the embedded fs that contains the implementations.
func NewContractImplementations(path string) (ContractImplementations, error) {
	var impls ContractImplementations
	data, err := implementationsFS.ReadFile(path)
	if err != nil {
		return impls, fmt.Errorf("failed to read implementations: %w", err)
	}
	if err := yaml.Unmarshal(data, &impls); err != nil {
		return impls, fmt.Errorf("failed to decode implementations: %w", err)
	}

	if impls.L1CrossDomainMessenger == nil {
		impls.L1CrossDomainMessenger = make(map[string]Address)
	}
	if impls.L1ERC721Bridge == nil {
		impls.L1ERC721Bridge = make(map[string]Address)
	}
	if impls.L1StandardBridge == nil {
		impls.L1StandardBridge = make(map[string]Address)
	}
	if impls.L2OutputOracle == nil {
		impls.L2OutputOracle = make(map[string]Address)
	}
	if impls.OptimismMintableERC20Factory == nil {
		impls.OptimismMintableERC20Factory = make(map[string]Address)
	}
	if impls.OptimismPortal == nil {
		impls.OptimismPortal = make(map[string]Address)
	}
	if impls.SystemConfig == nil {
		impls.SystemConfig = make(map[string]Address)
	}
	return impls, nil
}

// copySemverMap is a concrete implementation of maps.Copy for map[string]Address.
var copySemverMap = maps.Copy[map[string]Address, map[string]Address]

// Merge will combine two ContractImplementations into one. Any conflicting keys will
// be overwritten by the arguments. It assumes that nonce of the struct fields are nil.
func (c ContractImplementations) Merge(other ContractImplementations) {
	copySemverMap(c.L1CrossDomainMessenger, other.L1CrossDomainMessenger)
	copySemverMap(c.L1ERC721Bridge, other.L1ERC721Bridge)
	copySemverMap(c.L1StandardBridge, other.L1StandardBridge)
	copySemverMap(c.L2OutputOracle, other.L2OutputOracle)
	copySemverMap(c.OptimismMintableERC20Factory, other.OptimismMintableERC20Factory)
	copySemverMap(c.OptimismPortal, other.OptimismPortal)
	copySemverMap(c.SystemConfig, other.SystemConfig)
}

// Copy will return a shallow copy of the ContractImplementations.
func (c ContractImplementations) Copy() ContractImplementations {
	return ContractImplementations{
		L1CrossDomainMessenger:       maps.Clone(c.L1CrossDomainMessenger),
		L1ERC721Bridge:               maps.Clone(c.L1ERC721Bridge),
		L1StandardBridge:             maps.Clone(c.L1StandardBridge),
		L2OutputOracle:               maps.Clone(c.L2OutputOracle),
		OptimismMintableERC20Factory: maps.Clone(c.OptimismMintableERC20Factory),
		OptimismPortal:               maps.Clone(c.OptimismPortal),
		SystemConfig:                 maps.Clone(c.SystemConfig),
	}
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

var Implementations = map[uint64]ContractImplementations{}

func init() {
	// read the global implementations
	globalImpls, err := NewContractImplementations(path.Join("implementations", "implementations.yaml"))
	if err != nil {
		panic(fmt.Errorf("failed to read implementations: %w", err))
	}

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

		networkImpls, err := NewContractImplementations(path.Join("implementations", "networks", s.Name()+".yaml"))
		if err != nil {
			panic(fmt.Errorf("failed to read implementations of superchain target %s: %w", s.Name(), err))
		}

		implementations := globalImpls.Copy()
		implementations.Merge(networkImpls)

		chainID := superchainEntry.Config.L1.ChainID
		Implementations[chainID] = implementations
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
