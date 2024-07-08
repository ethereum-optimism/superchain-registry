package superchain

import (
	"compress/gzip"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
)

var ErrEmptyVersion = errors.New("empty version")

//go:embed configs
var superchainFS embed.FS

//go:embed extra/addresses extra/bytecodes extra/genesis extra/genesis-system-configs
var extraFS embed.FS

type BlockID struct {
	Hash   Hash   `yaml:"hash"`
	Number uint64 `yaml:"number"`
}

type ChainGenesis struct {
	L1           BlockID      `yaml:"l1"`
	L2           BlockID      `yaml:"l2"`
	L2Time       uint64       `json:"l2_time" yaml:"l2_time"`
	ExtraData    *HexBytes    `yaml:"extra_data,omitempty"`
	SystemConfig SystemConfig `json:"system_config" yaml:"-"`
}

type SystemConfig struct {
	BatcherAddr       Address `json:"batcherAddr"`
	Overhead          string  `json:"overhead"`
	Scalar            string  `json:"scalar"`
	GasLimit          uint64  `json:"gasLimit"`
	BaseFeeScalar     uint64  `json:"baseFeeScalar"`
	BlobBaseFeeScalar uint64  `json:"blobBaseFeeScalar"`
}

type GenesisData struct {
	L1     GenesisLayer `json:"l1" yaml:"l1"`
	L2     GenesisLayer `json:"l2" yaml:"l2"`
	L2Time int          `json:"l2_time" yaml:"l2_time"`
}

type GenesisLayer struct {
	Hash   string `json:"hash" yaml:"hash"`
	Number int    `json:"number" yaml:"number"`
}

type HardForkConfiguration struct {
	CanyonTime  *uint64 `json:"canyon_time,omitempty" yaml:"canyon_time,omitempty"`
	DeltaTime   *uint64 `json:"delta_time,omitempty" yaml:"delta_time,omitempty"`
	EcotoneTime *uint64 `json:"ecotone_time,omitempty" yaml:"ecotone_time,omitempty"`
	FjordTime   *uint64 `json:"fjord_time,omitempty" yaml:"fjord_time,omitempty"`
}

type SuperchainLevel uint

const (
	Standard SuperchainLevel = 2
	Frontier SuperchainLevel = 1
)

type ChainConfig struct {
	Name         string `yaml:"name"`
	ChainID      uint64 `yaml:"chain_id"`
	PublicRPC    string `yaml:"public_rpc"`
	SequencerRPC string `yaml:"sequencer_rpc"`
	Explorer     string `yaml:"explorer"`

	SuperchainLevel SuperchainLevel `yaml:"superchain_level"`

	// If StandardChainCandidate is true, standard chain validation checks will
	// run on this chain even if it is a frontier chain.
	StandardChainCandidate bool `yaml:"standard_chain_candidate,omitempty"`

	// If SuperchainTime is set, hardforks times after SuperchainTime
	// will be inherited from the superchain-wide config.
	SuperchainTime *uint64 `yaml:"superchain_time"`

	// An op-contracts tag from github.com/ethereum-optimism/optimism
	// from which contracts were deployed. Example: "op-contracts/v1.4.0".
	// May be nil for frontier chains.
	ContractsVersionTag *string `yaml:"contracts_version_tag"`

	BatchInboxAddr Address `yaml:"batch_inbox_addr"`

	Genesis ChainGenesis `yaml:"genesis"`

	// Superchain is a simple string to identify the superchain.
	// This is implied by directory structure, and not encoded in the config file itself.
	Superchain string `yaml:"-"`
	// Chain is a simple string to identify the chain, within its superchain context.
	// This matches the resource filename, it is not encoded in the config file itself.
	Chain string `yaml:"-"`

	// Hardfork Configuration Overrides
	HardForkConfiguration `yaml:",inline"`

	// Optional feature
	Plasma *PlasmaConfig `yaml:"plasma,omitempty"`
}

func (c ChainConfig) Identifier() string {
	return c.Superchain + "/" + c.Chain
}

type PlasmaConfig struct {
	DAChallengeAddress *Address `json:"da_challenge_contract_address" yaml:"-"`
	// DA challenge window value set on the DAC contract. Used in plasma mode
	// to compute when a commitment can no longer be challenged.
	DAChallengeWindow *uint64 `json:"da_challenge_window" yaml:"da_challenge_window"`
	// DA resolve window value set on the DAC contract. Used in plasma mode
	// to compute when a challenge expires and trigger a reorg if needed.
	DAResolveWindow *uint64 `json:"da_resolve_window" yaml:"da_resolve_window"`
}

// setNilHardforkTimestampsToDefaultOrZero overwrites each unspecified hardfork activation time override
// with the superchain default, if the default is not nil and is after the SuperchainTime. If the default
// is after the chain's l2 time, that hardfork activation time is set to zero (meaning "activates at genesis").
func (c *ChainConfig) setNilHardforkTimestampsToDefaultOrZero(s *SuperchainConfig) {
	if c.SuperchainTime == nil {
		// No changes if SuperchainTime is unset
		return
	}
	cVal := reflect.ValueOf(&c.HardForkConfiguration).Elem()
	sVal := reflect.ValueOf(&s.hardForkDefaults).Elem()

	var zero uint64 = 0
	ptrZero := reflect.ValueOf(&zero)

	// Iterate over hardfork timestamps (i.e. CanyontTime, DeltaTime, ...)
	for i := 0; i < reflect.Indirect(cVal).NumField(); i++ {
		overridePtr := cVal.Field(i)
		if !overridePtr.IsNil() {
			// No change if override is set
			continue
		}

		defaultPtr := sVal.Field(i)
		if defaultPtr.IsNil() {
			// No change if default is unset
			continue
		}

		defaultValue := reflect.Indirect(defaultPtr).Uint()
		if defaultValue < *c.SuperchainTime {
			// No change if hardfork activated before SuperchainTime
			continue
		}

		if defaultValue > c.Genesis.L2Time {
			// Use default value if is after genesis
			overridePtr.Set(defaultPtr)
		} else {
			// Use zero if it is equal to or before genesis
			overridePtr.Set(ptrZero)
		}
	}
}

// EnhanceYAML creates a customized yaml string from a RollupConfig. After completion,
// the *yaml.Node pointer can be used with a yaml encoder to write the custom format to file
func (c *ChainConfig) EnhanceYAML(ctx context.Context, node *yaml.Node) error {
	hexStringRegex := regexp.MustCompile(`^0x[a-fA-F0-9]+$`)

	// Check if context is done before processing
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0] // Dive into the document node
	}

	var lastKey string
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]
		if valNode.Kind == yaml.ScalarNode && valNode.Tag == "!!str" && hexStringRegex.MatchString(valNode.Value) {
			valNode.Style = yaml.DoubleQuotedStyle
		}

		// Add blank line AFTER these keys
		if lastKey == "explorer" || lastKey == "contracts_version_tag" || lastKey == "genesis" {
			keyNode.HeadComment = "\n"
		}

		if keyNode.Value == "contracts_version_tag" {
			switch valNode.Value {
			case "op-contracts/v1.3.0":
				keyNode.LineComment = "Multi-Chain Prep (MCP) https://github.com/ethereum-optimism/optimism/releases/tag/op-contracts%2Fv1.3.0"
			case "op-contracts/v1.4.0":
				keyNode.LineComment = "Fault Proofs https://github.com/ethereum-optimism/optimism/releases/tag/op-contracts%2Fv1.4.0"

			}
		}

		// Add blank line BEFORE these keys
		if keyNode.Value == "genesis" || keyNode.Value == "plasma" {
			keyNode.HeadComment = "\n"
		}

		// Recursive call to check nested fields for "_time" suffix
		if valNode.Kind == yaml.MappingNode {
			if err := c.EnhanceYAML(ctx, valNode); err != nil {
				return err
			}
		}

		if keyNode.Value == "superchain_time" {
			if valNode.Value == "" || valNode.Value == "null" {
				keyNode.LineComment = "Missing hardfork times are NOT yet inherited from superchain.yaml"
			} else if valNode.Value == "0" {
				keyNode.LineComment = "Missing hardfork times are inherited from superchain.yaml"
			} else {
				keyNode.LineComment = "Missing hardfork times after this time are inherited from superchain.yaml"
			}
		}

		// Add human readable timestamp in comment
		if strings.HasSuffix(keyNode.Value, "_time") && valNode.Value != "" && valNode.Value != "null" {
			t, err := strconv.ParseInt(valNode.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to convert yaml string timestamp to int: %w", err)
			}
			timestamp := time.Unix(t, 0).UTC()
			keyNode.LineComment = timestamp.Format("Mon 2 Jan 2006 15:04:05 UTC")
		}

		if keyNode.Value == "standard_chain_candidate" {
			keyNode.LineComment = "This is a temporary field which causes most of the standard validation checks to run on this chain"
		}

		lastKey = keyNode.Value
	}
	return nil
}

type Roles struct {
	SystemConfigOwner Address `json:"SystemConfigOwner"`
	ProxyAdminOwner   Address `json:"ProxyAdminOwner"`
	Guardian          Address `json:"Guardian"`
	Challenger        Address `json:"Challenger"`
	Proposer          Address `json:"Proposer"`
	UnsafeBlockSigner Address `json:"UnsafeBlockSigner"`
	BatchSubmitter    Address `json:"BatchSubmitter"`
}

// AddressList represents the set of network specific contracts and roles for a given network.
type AddressList struct {
	Roles                             `json:",inline"`
	AddressManager                    Address `json:"AddressManager"`
	L1CrossDomainMessengerProxy       Address `json:"L1CrossDomainMessengerProxy"`
	L1ERC721BridgeProxy               Address `json:"L1ERC721BridgeProxy"`
	L1StandardBridgeProxy             Address `json:"L1StandardBridgeProxy"`
	L2OutputOracleProxy               Address `json:"L2OutputOracleProxy"`
	OptimismMintableERC20FactoryProxy Address `json:"OptimismMintableERC20FactoryProxy"`
	OptimismPortalProxy               Address `json:"OptimismPortalProxy"`
	SystemConfigProxy                 Address `json:"SystemConfigProxy"`
	ProxyAdmin                        Address `json:"ProxyAdmin"`

	// Fault Proof contracts:
	AnchorStateRegistryProxy Address `json:"AnchorStateRegistryProxy,omitempty"`
	DelayedWETHProxy         Address `json:"DelayedWETHProxy,omitempty"`
	DisputeGameFactoryProxy  Address `json:"DisputeGameFactoryProxy,omitempty"`
	FaultDisputeGame         Address `json:"FaultDisputeGame,omitempty"`
	MIPS                     Address `json:"MIPS,omitempty"`
	PermissionedDisputeGame  Address `json:"PermissionedDisputeGame,omitempty"`
	PreimageOracle           Address `json:"PreimageOracle,omitempty"`
}

// AddressFor returns a nonzero address for the supplied name, if it has been specified
// (and an error otherwise). Useful for slicing into the struct using a string.
func (a AddressList) AddressFor(name string) (Address, error) {
	var address Address
	switch name {
	case "AddressManager":
		address = a.AddressManager
	case "ProxyAdmin":
		address = a.ProxyAdmin
	case "L1CrossDomainMessengerProxy":
		address = a.L1CrossDomainMessengerProxy
	case "L1ERC721BridgeProxy":
		address = a.L1ERC721BridgeProxy
	case "L1StandardBridgeProxy":
		address = a.L1StandardBridgeProxy
	case "L2OutputOracleProxy":
		address = a.L2OutputOracleProxy
	case "OptimismMintableERC20FactoryProxy":
		address = a.OptimismMintableERC20FactoryProxy
	case "OptimismPortalProxy":
		address = a.OptimismPortalProxy
	case "SystemConfigProxy":
		address = a.SystemConfigProxy
	case "AnchorStateRegistryProxy":
		address = a.AnchorStateRegistryProxy
	case "DelayedWETHProxy":
		address = a.DelayedWETHProxy
	case "DisputeGameFactoryProxy":
		address = a.DisputeGameFactoryProxy
	case "FaultDisputeGame":
		address = a.FaultDisputeGame
	case "MIPS":
		address = a.MIPS
	case "PermissionedDisputeGame":
		address = a.PermissionedDisputeGame
	case "PreimageOracle":
		address = a.PreimageOracle
	case "SystemConfigOwner":
		address = a.SystemConfigOwner
	case "ProxyAdminOwner":
		address = a.ProxyAdminOwner
	case "Guardian":
		address = a.Guardian
	case "Challenger":
		address = a.Challenger
	case "BatchSubmitter":
		address = a.BatchSubmitter
	case "UnsafeBlockSigner":
		address = a.UnsafeBlockSigner
	case "Proposer":
		address = a.Proposer
	default:
		return address, fmt.Errorf("no such name %s", name)
	}
	if address == (Address{}) {
		return address, fmt.Errorf("no address or zero address specified for  %s", name)
	}
	return address, nil
}

// AddressSet represents a set of addresses for a given
// contract. They are keyed by the semantic version.
type AddressSet map[string]Address

// VersionedContract represents a contract that has a semantic version.
type VersionedContract struct {
	Version string  `json:"version"`
	Address Address `json:"address"`
}

// ContractVersions represents the desired semantic version of the contracts
// in the superchain. This currently only supports L1 contracts but could
// represent L2 predeploys in the future.
type ContractVersions struct {
	L1CrossDomainMessenger       string `yaml:"l1_cross_domain_messenger" toml:"l1_cross_domain_messenger"`
	L1ERC721Bridge               string `yaml:"l1_erc721_bridge" toml:"l1_erc721_bridge"`
	L1StandardBridge             string `yaml:"l1_standard_bridge" toml:"l1_standard_bridge"`
	L2OutputOracle               string `yaml:"l2_output_oracle,omitempty" toml:"l2_output_oracle,omitempty"`
	OptimismMintableERC20Factory string `yaml:"optimism_mintable_erc20_factory" toml:"optimism_mintable_erc20_factory"`
	OptimismPortal               string `yaml:"optimism_portal" toml:"optimism_portal"`
	SystemConfig                 string `yaml:"system_config" toml:"system_config"`
	// Superchain-wide contracts:
	ProtocolVersions string `yaml:"protocol_versions" toml:"protocol_versions"`
	SuperchainConfig string `yaml:"superchain_config,omitempty"`
	// Fault Proof contracts:
	AnchorStateRegistry     string `yaml:"anchor_state_registry,omitempty" toml:"anchor_state_registry,omitempty"`
	DelayedWETH             string `yaml:"delayed_weth,omitempty" toml:"delayed_weth,omitempty"`
	DisputeGameFactory      string `yaml:"dispute_game_factory,omitempty" toml:"dispute_game_factory,omitempty"`
	FaultDisputeGame        string `yaml:"fault_dispute_game,omitempty" toml:"fault_dispute_game,omitempty"`
	MIPS                    string `yaml:"mips,omitempty" toml:"mips,omitempty"`
	PermissionedDisputeGame string `yaml:"permissioned_dispute_game,omitempty" toml:"permissioned_dispute_game,omitempty"`
	PreimageOracle          string `yaml:"preimage_oracle,omitempty" toml:"preimage_oracle,omitempty"`
}

// VersionFor returns the version for the supplied contract name, if it exits
// (and an error otherwise). Useful for slicing into the struct using a string.
func (c ContractVersions) VersionFor(contractName string) (string, error) {
	var version string
	switch contractName {
	case "L1CrossDomainMessenger":
		version = c.L1CrossDomainMessenger
	case "L1ERC721Bridge":
		version = c.L1ERC721Bridge
	case "L1StandardBridge":
		version = c.L1StandardBridge
	case "L2OutputOracle":
		version = c.L2OutputOracle
	case "OptimismMintableERC20Factory":
		version = c.OptimismMintableERC20Factory
	case "OptimismPortal":
		version = c.OptimismPortal
	case "SystemConfig":
		version = c.SystemConfig
	case "AnchorStateRegistry":
		version = c.AnchorStateRegistry
	case "DelayedWETH":
		version = c.DelayedWETH
	case "DisputeGameFactory":
		version = c.DisputeGameFactory
	case "FaultDisputeGame":
		version = c.FaultDisputeGame
	case "MIPS":
		version = c.MIPS
	case "PermissionedDisputeGame":
		version = c.PermissionedDisputeGame
	case "PreimageOracle":
		version = c.PreimageOracle
	case "ProtocolVersions":
		version = c.ProtocolVersions
	case "SuperchainConfig":
		version = c.SuperchainConfig
	default:
		return "", errors.New("no such contract name")
	}
	if version == "" {
		return "", errors.New("no version specified")
	}
	return version, nil
}

// Check will sanity check the validity of the semantic version strings
// in the ContractVersions struct. If allowEmptyVersions is true, empty version errors will be ignored.
func (c ContractVersions) Check(allowEmptyVersions bool) error {
	val := reflect.ValueOf(c)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		str, ok := field.Interface().(string)
		if !ok {
			return fmt.Errorf("invalid type for field %s", val.Type().Field(i).Name)
		}
		if str == "" {
			if allowEmptyVersions {
				continue // we allow empty strings and rely on tests to assert (or except) a nonempty version
			}
			return fmt.Errorf("empty version for field %s", val.Type().Field(i).Name)
		}
		str = canonicalizeSemver(str)
		if !semver.IsValid(str) {
			return fmt.Errorf("invalid semver %s for field %s", str, val.Type().Field(i).Name)
		}
	}
	return nil
}

// canonicalizeSemver will ensure that the version string has a "v" prefix.
// This is because the semver library being used requires the "v" prefix,
// even though
func canonicalizeSemver(version string) string {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
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
	Nonce         uint64  `json:"nonce"`
	Timestamp     uint64  `json:"timestamp"`
	ExtraData     []byte  `json:"extraData"`
	GasLimit      uint64  `json:"gasLimit"`
	Difficulty    *HexBig `json:"difficulty"`
	Mixhash       Hash    `json:"mixHash"`
	Coinbase      Address `json:"coinbase"`
	Number        uint64  `json:"number"`
	GasUsed       uint64  `json:"gasUsed"`
	ParentHash    Hash    `json:"parentHash"`
	BaseFee       *HexBig `json:"baseFeePerGas"`
	ExcessBlobGas *uint64 `json:"excessBlobGas"` // EIP-4844
	BlobGasUsed   *uint64 `json:"blobGasUsed"`   // EIP-4844
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

	ProtocolVersionsAddr *Address `yaml:"protocol_versions_addr,omitempty"`
	SuperchainConfigAddr *Address `yaml:"superchain_config_addr,omitempty"`

	// Hardfork Configuration. These values may be overridden by individual chains.
	hardForkDefaults HardForkConfiguration
}

// custom unmarshal function to allow yaml to be unmarshalled into unexported fields
func unMarshalSuperchainConfig(data []byte, s *SuperchainConfig) error {
	temp := struct {
		*SuperchainConfig `yaml:",inline"`
		HardForks         *HardForkConfiguration `yaml:",inline"`
	}{
		SuperchainConfig: s,
		HardForks:        &s.hardForkDefaults,
	}

	return yaml.Unmarshal(data, temp)
}

type Superchain struct {
	Config SuperchainConfig

	// Chains that are part of this superchain
	ChainIDs []uint64

	// Superchain identifier, without capitalization or display changes.
	Superchain string
}

// IsEcotone returns true if the EcotoneTime for this chain in the past.
func (c *ChainConfig) IsEcotone() bool {
	if et := c.EcotoneTime; et != nil {
		return int64(*et) < time.Now().Unix()
	}
	return false
}

var Superchains = map[string]*Superchain{}

var OPChains = map[uint64]*ChainConfig{}

var Addresses = map[uint64]*AddressList{}

var GenesisSystemConfigs = map[uint64]*GenesisSystemConfig{}

// SuperchainSemver maps superchain name to a contract name : approved semver version structure.
var SuperchainSemver map[string]ContractVersions

func isConfigFile(c fs.DirEntry) bool {
	return (!c.IsDir() &&
		strings.HasSuffix(c.Name(), ".yaml") &&
		c.Name() != "superchain.yaml" &&
		c.Name() != "semver.yaml")
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
