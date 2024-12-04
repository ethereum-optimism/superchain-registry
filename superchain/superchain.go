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
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"golang.org/x/mod/semver"
)

var ErrEmptyVersion = errors.New("empty version")

//go:embed configs
var superchainFS embed.FS

//go:embed extra/bytecodes extra/genesis
var extraFS embed.FS

type BlockID struct {
	Hash   Hash   `json:"hash" toml:"hash"`
	Number uint64 `json:"number" toml:"number"`
}

type OptimismConfig struct {
	EIP1559Elasticity        uint64  `toml:"eip1559_elasticity" json:"eip1559Elasticity"`
	EIP1559Denominator       uint64  `toml:"eip1559_denominator" json:"eip1559Denominator"`
	EIP1559DenominatorCanyon *uint64 `toml:"eip1559_denominator_canyon,omitempty" json:"eip1559DenominatorCanyon,omitempty"`
}

type ChainGenesis struct {
	L1           BlockID      `json:"l1" toml:"l1"`
	L2           BlockID      `json:"l2" toml:"l2"`
	L2Time       uint64       `json:"l2_time" toml:"l2_time"`
	ExtraData    *HexBytes    `json:"extra_data,omitempty" toml:"extra_data,omitempty"`
	SystemConfig SystemConfig `json:"system_config" toml:"system_config"`
}

type SystemConfig struct {
	BatcherAddr       Address `json:"batcherAddr" toml:"batcherAddress"`
	Overhead          Hash    `json:"overhead" toml:"overhead"`
	Scalar            Hash    `json:"scalar" toml:"scalar"`
	GasLimit          uint64  `json:"gasLimit" toml:"gasLimit"`
	BaseFeeScalar     *uint64 `json:"baseFeeScalar,omitempty" toml:"baseFeeScalar,omitempty"`
	BlobBaseFeeScalar *uint64 `json:"blobBaseFeeScalar,omitempty" toml:"blobBaseFeeScalar,omitempty"`
}

type HardForkConfiguration struct {
	CanyonTime   *uint64 `json:"canyon_time,omitempty" toml:"canyon_time,omitempty"`
	DeltaTime    *uint64 `json:"delta_time,omitempty" toml:"delta_time,omitempty"`
	EcotoneTime  *uint64 `json:"ecotone_time,omitempty" toml:"ecotone_time,omitempty"`
	FjordTime    *uint64 `json:"fjord_time,omitempty" toml:"fjord_time,omitempty"`
	GraniteTime  *uint64 `json:"granite_time,omitempty" toml:"granite_time,omitempty"`
	HoloceneTime *uint64 `json:"holocene_time,omitempty" toml:"holocene_time,omitempty"`
	IsthmusTime  *uint64 `json:"isthmus_time,omitempty" toml:"isthmus_time,omitempty"`
}

type SuperchainLevel uint

const (
	Standard SuperchainLevel = 1
	Frontier SuperchainLevel = 0
)

type DataAvailability string

const (
	EthDA DataAvailability = "eth-da"
	AltDA DataAvailability = "alt-da"
)

type ChainConfig struct {
	Name         string `toml:"name"`
	ChainID      uint64 `toml:"chain_id" json:"l2_chain_id"`
	PublicRPC    string `toml:"public_rpc"`
	SequencerRPC string `toml:"sequencer_rpc"`
	Explorer     string `toml:"explorer"`

	SuperchainLevel SuperchainLevel `toml:"superchain_level"`

	// If StandardChainCandidate is true, standard chain validation checks will
	// run on this chain even if it is a frontier chain.
	StandardChainCandidate bool `toml:"standard_chain_candidate,omitempty"`

	// If SuperchainTime is set, hardforks times after SuperchainTime
	// will be inherited from the superchain-wide config.
	SuperchainTime *uint64 `toml:"superchain_time"`

	BatchInboxAddr Address `toml:"batch_inbox_addr" json:"batch_inbox_address"`

	// Superchain is a simple string to identify the superchain.
	// This is implied by directory structure, and not encoded in the config file itself.
	Superchain string `toml:"-"`
	// Chain is a simple string to identify the chain, within its superchain context.
	// This matches the resource filename, it is not encoded in the config file itself.
	Chain string `toml:"-"`

	// Hardfork Configuration Overrides
	HardForkConfiguration `toml:",inline"`

	BlockTime            uint64           `toml:"block_time" json:"block_time"`
	SequencerWindowSize  uint64           `toml:"seq_window_size" json:"seq_window_size"`
	MaxSequencerDrift    uint64           `toml:"max_sequencer_drift" json:"max_sequencer_drift"`
	DataAvailabilityType DataAvailability `toml:"data_availability_type"`
	Optimism             *OptimismConfig  `toml:"optimism,omitempty" json:"optimism,omitempty"`

	// Optional feature
	AltDA *AltDAConfig `toml:"alt_da,omitempty" json:"alt_da,omitempty"`

	GasPayingToken *Address `toml:"gas_paying_token,omitempty"` // Just metadata, not consumed by downstream OPStack software

	Genesis ChainGenesis `toml:"genesis" json:"genesis"`

	Addresses AddressList `toml:"addresses"`
}

func (c ChainConfig) Identifier() string {
	return c.Superchain + "/" + c.Chain
}

// Returns a shallow copy of the chain config with some fields mutated
// to declare the chain a standard chain. No fields on the receiver
// are mutated.
func (c *ChainConfig) PromoteToStandard() (*ChainConfig, error) {
	if !c.StandardChainCandidate {
		return nil, errors.New("can only promote standard candidate chains")
	}
	if c.SuperchainLevel != Frontier {
		return nil, errors.New("can only promote frontier chains")
	}

	// Note that any pointers in c are copied to d
	// This is not problematic as long as we do
	// not modify the values pointed to.
	d := *c

	d.StandardChainCandidate = false
	d.SuperchainLevel = Standard
	now := uint64(time.Now().Unix())
	d.SuperchainTime = &now
	return &d, nil
}

type AltDAConfig struct {
	DAChallengeAddress *Address `json:"da_challenge_contract_address" toml:"da_challenge_contract_address"`
	// DA challenge window value set on the DAC contract. Used in altDA mode
	// to compute when a commitment can no longer be challenged.
	DAChallengeWindow *uint64 `json:"da_challenge_window" toml:"da_challenge_window"`
	// DA resolve window value set on the DAC contract. Used in altDA mode
	// to compute when a challenge expires and trigger a reorg if needed.
	DAResolveWindow  *uint64 `json:"da_resolve_window" toml:"da_resolve_window"`
	DACommitmentType *string `json:"da_commitment_type" toml:"da_commitment_type"`
}

func (c *ChainConfig) CheckDataAvailability() error {
	c.DataAvailabilityType = EthDA
	if c.AltDA != nil {
		c.DataAvailabilityType = AltDA
		if c.AltDA.DAChallengeAddress == nil {
			return fmt.Errorf("missing required altDA field: da_challenge_contract_address")
		}
	}

	return nil
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

// MarshalJSON excludes any addresses set to 0x000...000
func (a AddressList) MarshalJSON() ([]byte, error) {
	type AddressList2 AddressList // use another type to prevent infinite recursion later on
	b := AddressList2(a)

	o, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	out := make(map[string]Address)
	err = json.Unmarshal(o, &out)
	if err != nil {
		return nil, err
	}

	for k, v := range out {
		if (v == Address{}) {
			delete(out, k)
		}
	}

	return json.Marshal(out)
}

func (c *ChainConfig) GenerateTOMLComments(ctx context.Context) (map[string]string, error) {
	comments := make(map[string]string)

	createTimestampComment := func(fieldName string, fieldValue *uint64, comments map[string]string) {
		if fieldValue != nil {
			timestamp := time.Unix(int64(*fieldValue), 0).UTC()
			comments[fieldName] = fmt.Sprintf("# %s", timestamp.Format("Mon 2 Jan 2006 15:04:05 UTC"))
		}
	}

	if c.SuperchainTime != nil {
		if *c.SuperchainTime == 0 {
			comments["superchain_time"] = "# Missing hardfork times are inherited from superchain.toml"
		} else {
			createTimestampComment("superchain_time", c.SuperchainTime, comments)
		}
	}

	createTimestampComment("canyon_time", c.CanyonTime, comments)
	createTimestampComment("delta_time", c.DeltaTime, comments)
	createTimestampComment("ecotone_time", c.EcotoneTime, comments)
	createTimestampComment("fjord_time", c.FjordTime, comments)
	createTimestampComment("granite_time", c.GraniteTime, comments)
	createTimestampComment("holocene_time", c.HoloceneTime, comments)
	createTimestampComment("isthmus_time", c.IsthmusTime, comments)

	if c.StandardChainCandidate {
		comments["standard_chain_candidate"] = "# This is a temporary field which causes most of the standard validation checks to run on this chain"
	}

	return comments, nil
}

type Roles struct {
	SystemConfigOwner Address `json:"SystemConfigOwner" toml:"SystemConfigOwner"`
	ProxyAdminOwner   Address `json:"ProxyAdminOwner" toml:"ProxyAdminOwner"`
	Guardian          Address `json:"Guardian" toml:"Guardian"`
	Challenger        Address `json:"Challenger" toml:"Challenger"`
	Proposer          Address `json:"Proposer" toml:"Proposer"`
	UnsafeBlockSigner Address `json:"UnsafeBlockSigner" toml:"UnsafeBlockSigner"`
	BatchSubmitter    Address `json:"BatchSubmitter" toml:"BatchSubmitter"`
}

// AddressList represents the set of network specific contracts and roles for a given network.
type AddressList struct {
	Roles                             `json:",inline" toml:",inline"`
	AddressManager                    Address `json:"AddressManager" toml:"AddressManager"`
	L1CrossDomainMessengerProxy       Address `json:"L1CrossDomainMessengerProxy" toml:"L1CrossDomainMessengerProxy"`
	L1ERC721BridgeProxy               Address `json:"L1ERC721BridgeProxy" toml:"L1ERC721BridgeProxy"`
	L1StandardBridgeProxy             Address `json:"L1StandardBridgeProxy" toml:"L1StandardBridgeProxy"`
	L2OutputOracleProxy               Address `json:"L2OutputOracleProxy" toml:"L2OutputOracleProxy,omitempty"`
	OptimismMintableERC20FactoryProxy Address `json:"OptimismMintableERC20FactoryProxy" toml:"OptimismMintableERC20FactoryProxy"`
	OptimismPortalProxy               Address `json:"OptimismPortalProxy,omitempty" toml:"OptimismPortalProxy,omitempty"`
	SystemConfigProxy                 Address `json:"SystemConfigProxy" toml:"SystemConfigProxy"`
	ProxyAdmin                        Address `json:"ProxyAdmin" toml:"ProxyAdmin"`
	SuperchainConfig                  Address `json:"SuperchainConfig,omitempty" toml:"SuperchainConfig,omitempty"`

	// Fault Proof contracts:
	AnchorStateRegistryProxy Address `json:"AnchorStateRegistryProxy,omitempty" toml:"AnchorStateRegistryProxy,omitempty"`
	DelayedWETHProxy         Address `json:"DelayedWETHProxy,omitempty" toml:"DelayedWETHProxy,omitempty"`
	DisputeGameFactoryProxy  Address `json:"DisputeGameFactoryProxy,omitempty" toml:"DisputeGameFactoryProxy,omitempty"`
	FaultDisputeGame         Address `json:"FaultDisputeGame,omitempty" toml:"FaultDisputeGame,omitempty"`
	MIPS                     Address `json:"MIPS,omitempty" toml:"MIPS,omitempty"`
	PermissionedDisputeGame  Address `json:"PermissionedDisputeGame,omitempty" toml:"PermissionedDisputeGame,omitempty"`
	PreimageOracle           Address `json:"PreimageOracle,omitempty" toml:"PreimageOracle,omitempty"`

	// AltDA contracts:
	DAChallengeAddress Address `json:"DAChallengeAddress,omitempty" toml:"DAChallengeAddress,omitempty"`
}

// AddressFor returns a nonzero address for the supplied name, if it has been specified
// (and an error otherwise).
func (a AddressList) AddressFor(name string) (Address, error) {
	// Use reflection to get the struct value and type
	v := reflect.ValueOf(a)

	// Try to find the field by name
	field := v.FieldByName(name)
	if !field.IsValid() {
		return Address{}, fmt.Errorf("no such name %s", name)
	}

	// Check if the field is of type Address
	if field.Type() != reflect.TypeOf(Address{}) {
		return Address{}, fmt.Errorf("field %s is not of type Address", name)
	}

	// Check if the address is a non-zero value
	address := field.Interface().(Address)
	if address == (Address{}) {
		return Address{}, fmt.Errorf("no address or zero address specified for %s", name)
	}

	return address, nil
}

type MappedContractProperties[T string | VersionedContract] struct {
	L1CrossDomainMessenger       T `toml:"l1_cross_domain_messenger,omitempty"`
	L1ERC721Bridge               T `toml:"l1_erc721_bridge,omitempty"`
	L1StandardBridge             T `toml:"l1_standard_bridge,omitempty"`
	L2OutputOracle               T `toml:"l2_output_oracle,omitempty"`
	OptimismMintableERC20Factory T `toml:"optimism_mintable_erc20_factory,omitempty"`
	OptimismPortal               T `toml:"optimism_portal,omitempty"`
	OptimismPortal2              T `toml:"optimism_portal2,omitempty"`
	SystemConfig                 T `toml:"system_config,omitempty"`
	// Superchain-wide contracts:
	ProtocolVersions T `toml:"protocol_versions,omitempty"`
	SuperchainConfig T `toml:"superchain_config,omitempty"`
	// Fault Proof contracts:
	AnchorStateRegistry     T `toml:"anchor_state_registry,omitempty"`
	DelayedWETH             T `toml:"delayed_weth,omitempty"`
	DisputeGameFactory      T `toml:"dispute_game_factory,omitempty"`
	FaultDisputeGame        T `toml:"fault_dispute_game,omitempty"`
	MIPS                    T `toml:"mips,omitempty"`
	PermissionedDisputeGame T `toml:"permissioned_dispute_game,omitempty"`
	PreimageOracle          T `toml:"preimage_oracle,omitempty"`
	CannonFaultDisputeGame  T `toml:"cannon_fault_dispute_game,omitempty"`
}

// ContractBytecodeHashes stores a bytecode hash against each contract
type ContractBytecodeHashes MappedContractProperties[string]

// VersionedContract represents a contract that has a semantic version.
type VersionedContract struct {
	Version string `toml:"version"`
	// If the contract is a superchain singleton, it will have a static address
	Address *Address `toml:"address,omitempty"`
	// If the contract is proxied, the implementation will have a static address
	ImplementationAddress *Address `toml:"implementation_address,omitempty"`
}

// ContractVersions represents the desired semantic version of the contracts
// in the superchain. This currently only supports L1 contracts but could
// represent L2 predeploys in the future.
type ContractVersions MappedContractProperties[VersionedContract]

// GetNonEmpty returns a slice of contract names, with an entry for each contract
// in the receiver with a non empty Version property.
func (c ContractVersions) GetNonEmpty() []string {
	// Get the value and type of the struct
	v := reflect.ValueOf(c)
	t := reflect.TypeOf(c)

	var fieldNames []string

	// Iterate through the struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Ensure the field is of type VersionedContract
		if field.Type() == reflect.TypeOf(VersionedContract{}) {
			// Get the Version field from the VersionedContract
			versionField := field.FieldByName("Version")

			// Check if the Version is non-empty
			if versionField.IsValid() && versionField.String() != "" {
				fieldNames = append(fieldNames, fieldType.Name)
			}
		}
	}

	return fieldNames
}

// VersionFor returns the version for the supplied contract name, if it exits
// (and an error otherwise). Useful for slicing into the struct using a string.
func (c ContractVersions) VersionFor(contractName string) (string, error) {
	// Use reflection to get the value of the struct
	val := reflect.ValueOf(c)
	// Get the field by name (contractName)
	field := val.FieldByName(contractName)

	// Check if the field exists and is a struct
	if !field.IsValid() {
		return "", errors.New("no such contract name")
	}

	// Check if the struct contains the "Version" field
	versionField := field.FieldByName("Version")
	if !versionField.IsValid() || versionField.String() == "" {
		return "", errors.New("no version specified")
	}

	// Return the version if it's a string
	if versionField.Kind() == reflect.String {
		return versionField.String(), nil
	}

	return "", errors.New("version is not a string")
}

// Check will sanity check the validity of the semantic version strings
// in the ContractVersions struct. If allowEmptyVersions is true, empty version errors will be ignored.
func (c ContractVersions) Check(allowEmptyVersions bool) error {
	val := reflect.ValueOf(c)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		vC, ok := field.Interface().(VersionedContract)
		if !ok {
			return fmt.Errorf("invalid type for field %s", val.Type().Field(i).Name)
		}
		if vC.Version == "" {
			if allowEmptyVersions {
				continue // we allow empty strings and rely on tests to assert (or except) a nonempty version
			}
			return fmt.Errorf("empty version for field %s", val.Type().Field(i).Name)
		}
		vC.Version = CanonicalizeSemver(vC.Version)
		if !semver.IsValid(vC.Version) {
			return fmt.Errorf("invalid semver %s for field %s", vC.Version, val.Type().Field(i).Name)
		}
	}
	return nil
}

// CanonicalizeSemver will ensure that the version string has a "v" prefix.
// This is because the semver library being used requires the "v" prefix,
// even though
func CanonicalizeSemver(version string) string {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
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
	ChainID   uint64 `toml:"chain_id"`
	PublicRPC string `toml:"public_rpc"`
	Explorer  string `toml:"explorer"`
}

type SuperchainConfig struct {
	Name string           `toml:"name"`
	L1   SuperchainL1Info `toml:"l1"`

	ProtocolVersionsAddr        *Address `toml:"protocol_versions_addr,omitempty"`
	SuperchainConfigAddr        *Address `toml:"superchain_config_addr,omitempty"`
	OPContractsManagerProxyAddr *Address `toml:"op_contracts_manager_proxy_addr,omitempty"`

	// Hardfork Configuration. These values may be overridden by individual chains.
	hardForkDefaults HardForkConfiguration
}

// custom unmarshal function to allow toml to be unmarshalled into unexported fields
func unMarshalSuperchainConfig(data []byte, s *SuperchainConfig) error {
	temp := struct {
		*SuperchainConfig      `toml:",inline"`
		*HardForkConfiguration `toml:",inline"`
	}{
		s,
		&s.hardForkDefaults,
	}

	return toml.Unmarshal(data, &temp)
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

var GenesisSystemConfigs = map[uint64]*SystemConfig{}

func isConfigFile(c fs.DirEntry) bool {
	return (!c.IsDir() &&
		strings.HasSuffix(c.Name(), ".toml") &&
		c.Name() != "superchain.toml" &&
		c.Name() != "semver.toml")
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
