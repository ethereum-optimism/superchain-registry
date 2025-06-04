package state

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer/opaque_map"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/superchain"
	"github.com/hashicorp/go-multierror"
	"github.com/tomwright/dasel"
)

//go:embed configs/v1-state.json
var standardV1State []byte

//go:embed configs/v1-intent.toml
var standardV1Intent []byte

//go:embed configs/v2-state.json
var standardV2State []byte

//go:embed configs/v2-intent.toml
var standardV2Intent []byte

//go:embed configs/v3-state.json
var standardV3State []byte

//go:embed configs/v3-intent.toml
var standardV3Intent []byte

//go:embed configs/v4-state.json
var standardV4State []byte

//go:embed configs/v4-intent.toml
var standardV4Intent []byte

func ReadOpaqueStateFile(p string) (OpaqueState, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer f.Close()

	var out OpaqueState
	if err := json.NewDecoder(f).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return out, nil
}

type StateMerger = func(state OpaqueState) (opaque_map.OpaqueMap, OpaqueState, error)

func GetStateMerger(version string) (StateMerger, error) {
	// Extract the version number using regex
	re := regexp.MustCompile(`op-deployer/v\d+\.(\d+)\.\d+`)
	match := re.FindStringSubmatch(version)

	if len(match) < 2 {
		return nil, fmt.Errorf("invalid deployer version format: %s", version)
	}

	// Get the middle version number
	versionNum, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse version number: %w", err)
	}

	// Return the appropriate merge function
	switch versionNum {
	case 0, 1:
		return MergeStateV1, nil
	case 2:
		return MergeStateV2, nil
	case 3:
		return MergeStateV3, nil
	case 4:
		return MergeStateV4, nil
	default:
		return nil, fmt.Errorf("unsupported deployer version: %d", versionNum)
	}
}

func MergeStateV1(userState OpaqueState) (opaque_map.OpaqueMap, OpaqueState, error) {
	l1ChainID, err := userState.ReadL1ChainID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read L1 chain ID: %w", err)
	}
	stdIntent, err := StandardIntentV1(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard intent: %w", err)
	}
	stdState, err := StandardStateV1(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard state: %w", err)
	}
	return mergeStateV2(userState, stdIntent, stdState)
}

func MergeStateV2(userState OpaqueState) (opaque_map.OpaqueMap, OpaqueState, error) {
	l1ChainID, err := userState.ReadL1ChainID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read L1 chain ID: %w", err)
	}
	stdIntent, err := StandardIntentV2(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard intent: %w", err)
	}
	stdState, err := StandardStateV2(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard state: %w", err)
	}
	return mergeStateV2(userState, stdIntent, stdState)
}

func MergeStateV3(userState OpaqueState) (opaque_map.OpaqueMap, OpaqueState, error) {
	l1ChainID, err := userState.ReadL1ChainID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read L1 chain ID: %w", err)
	}
	stdIntent, err := StandardIntentV3(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard intent: %w", err)
	}
	stdState, err := StandardStateV3(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard state: %w", err)
	}
	// V2 is correct here. V3's state is the same as V2, except with a
	// slightly different intent that contains operator fee fields.
	return mergeStateV2(userState, stdIntent, stdState)
}

func MergeStateV4(userState OpaqueState) (opaque_map.OpaqueMap, OpaqueState, error) {
	l1ChainID, err := userState.ReadL1ChainID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read L1 chain ID: %w", err)
	}
	stdIntent, err := StandardIntentV4(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard intent: %w", err)
	}
	stdState, err := StandardStateV4(l1ChainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create standard state: %w", err)
	}
	return mergeStateV4(userState, stdIntent, stdState)
}

func mergeStateV4(userState OpaqueState, stdIntent opaque_map.OpaqueMap, stdState OpaqueState) (opaque_map.OpaqueMap, OpaqueState, error) {
	userStateNode := dasel.New(userState)
	stdIntentNode := dasel.New(stdIntent)
	stdStateNode := dasel.New(stdState)

	appliedIntentNode, err := userStateNode.Query("appliedIntent")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read applied intent: %w", err)
	}

	// Helper function to aggregate copy errors
	var copyErrs error
	guard := func(err error) {
		if err != nil {
			copyErrs = multierror.Append(copyErrs, fmt.Errorf("failed to copy value: %w", err))
		}
	}

	// First, set up the intent
	guard(copyValue(appliedIntentNode, stdIntentNode, "l1ChainID"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].id"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].baseFeeVaultRecipient"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].l1FeeVaultRecipient"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].sequencerFeeVaultRecipient"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.l1ProxyAdminOwner"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.l2ProxyAdminOwner"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.systemConfigOwner"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.unsafeBlockSigner"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.batcher"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.proposer"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.challenger"))

	// Then, set up the state
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].id"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].OpChainProxyAdminImpl"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].AddressManagerImpl"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].L1Erc721BridgeProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].SystemConfigProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].OptimismMintableErc20FactoryProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].L1StandardBridgeProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].L1CrossDomainMessengerProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].OptimismPortalProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].DisputeGameFactoryProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].AnchorStateRegistryProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].FaultDisputeGameImpl"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].PermissionedDisputeGameImpl"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].DelayedWethPermissionedGameProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].DelayedWethPermissionlessGameProxy"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].startBlock"))

	if copyErrs != nil {
		return nil, nil, copyErrs
	}

	intentResult, okIntent := stdIntentNode.InterfaceValue().(opaque_map.OpaqueMap)
	if !okIntent {
		return nil, nil, fmt.Errorf("internal error: synthesized intent is not OpaqueMapping, but %T", stdIntentNode.InterfaceValue())
	}
	stateResult, okState := stdStateNode.InterfaceValue().(OpaqueState)
	if !okState {
		return nil, nil, fmt.Errorf("internal error: synthesized state is not OpaqueMapping, but %T", stdStateNode.InterfaceValue())
	}

	return intentResult, stateResult, nil
}

func mergeStateV2(userState OpaqueState, stdIntent opaque_map.OpaqueMap, stdState OpaqueState) (opaque_map.OpaqueMap, OpaqueState, error) {
	userStateNode := dasel.New(userState)
	stdIntentNode := dasel.New(stdIntent)
	stdStateNode := dasel.New(stdState)

	appliedIntentNode, err := userStateNode.Query("appliedIntent")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read applied intent: %w", err)
	}

	// Helper function to aggregate copy errors
	var copyErrs error
	guard := func(err error) {
		if err != nil {
			copyErrs = multierror.Append(copyErrs, fmt.Errorf("failed to copy value: %w", err))
		}
	}

	// First, set up the intent
	guard(copyValue(appliedIntentNode, stdIntentNode, "l1ChainID"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].id"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].baseFeeVaultRecipient"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].l1FeeVaultRecipient"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].sequencerFeeVaultRecipient"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.l1ProxyAdminOwner"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.l2ProxyAdminOwner"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.systemConfigOwner"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.unsafeBlockSigner"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.batcher"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.proposer"))
	guard(copyValue(appliedIntentNode, stdIntentNode, "chains.[0].roles.challenger"))

	// Then, set up the state
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].id"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].proxyAdminAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].addressManagerAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].l1ERC721BridgeProxyAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].systemConfigProxyAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].optimismMintableERC20FactoryProxyAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].l1StandardBridgeProxyAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].l1CrossDomainMessengerProxyAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].optimismPortalProxyAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].disputeGameFactoryProxyAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].anchorStateRegistryProxyAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].faultDisputeGameAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].permissionedDisputeGameAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].delayedWETHPermissionedGameProxyAddress"))
	guard(copyValue(userStateNode, stdStateNode, "opChainDeployments.[0].startBlock"))

	if copyErrs != nil {
		return nil, nil, copyErrs
	}

	intentResult, okIntent := stdIntentNode.InterfaceValue().(opaque_map.OpaqueMap)
	if !okIntent {
		return nil, nil, fmt.Errorf("internal error: synthesized intent is not OpaqueMapping, but %T", stdIntentNode.InterfaceValue())
	}
	stateResult, okState := stdStateNode.InterfaceValue().(OpaqueState)
	if !okState {
		return nil, nil, fmt.Errorf("internal error: synthesized state is not OpaqueMapping, but %T", stdStateNode.InterfaceValue())
	}

	return intentResult, stateResult, nil
}

func StandardIntentV4(l1ChainID uint64) (opaque_map.OpaqueMap, error) {
	return standardIntentV4(l1ChainID, standardV4Intent)
}

func StandardIntentV3(l1ChainID uint64) (opaque_map.OpaqueMap, error) {
	return standardIntentV2(l1ChainID, standardV3Intent)
}

func StandardIntentV2(l1ChainID uint64) (opaque_map.OpaqueMap, error) {
	return standardIntentV2(l1ChainID, standardV2Intent)
}

func StandardIntentV1(l1ChainID uint64) (opaque_map.OpaqueMap, error) {
	return standardIntentV1(l1ChainID, standardV1Intent)
}

func standardIntentV4(l1ChainID uint64, data []byte) (opaque_map.OpaqueMap, error) {
	intent := make(opaque_map.OpaqueMap)
	if err := toml.Unmarshal(data, &intent); err != nil {
		panic(err)
	}

	var stdRoles validation.RolesConfig
	switch l1ChainID {
	case 1:
		stdRoles = validation.StandardConfigRolesMainnet
	case 11155111:
		stdRoles = validation.StandardConfigRolesSepolia
	default:
		return nil, fmt.Errorf("unsupported L1 chain ID: %d", l1ChainID)
	}

	root := dasel.New(intent)
	mustPutString(root, "superchainRoles.SuperchainProxyAdminOwner", stdRoles.L1ProxyAdminOwner)
	mustPutString(root, "superchainRoles.ProtocolVersionsOwner", stdRoles.ProtocolVersionsOwner)
	mustPutString(root, "superchainRoles.SuperchainGuardian", stdRoles.Guardian)

	return intent, nil
}

func standardIntentV2(l1ChainID uint64, data []byte) (opaque_map.OpaqueMap, error) {
	intent := make(opaque_map.OpaqueMap)
	if err := toml.Unmarshal(data, &intent); err != nil {
		panic(err)
	}

	var stdRoles validation.RolesConfig
	switch l1ChainID {
	case 1:
		stdRoles = validation.StandardConfigRolesMainnet
	case 11155111:
		stdRoles = validation.StandardConfigRolesSepolia
	default:
		return nil, fmt.Errorf("unsupported L1 chain ID: %d", l1ChainID)
	}

	root := dasel.New(intent)
	mustPutString(root, "superchainRoles.proxyAdminOwner", stdRoles.L1ProxyAdminOwner)
	mustPutString(root, "superchainRoles.protocolVersionsOwner", stdRoles.ProtocolVersionsOwner)
	mustPutString(root, "superchainRoles.guardian", stdRoles.Guardian)

	return intent, nil
}

// Add this type near the top of the file
type stringWrapper string

func (s stringWrapper) String() string {
	return string(s)
}

func standardIntentV1(l1ChainID uint64, data []byte) (opaque_map.OpaqueMap, error) {
	intent, err := standardIntentV2(l1ChainID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create standard intent: %w", err)
	}

	root := dasel.New(intent)
	// This is a hack to workaround an op-deployer bug where the protocolVersionsOwner is incorrectly
	// set to the protocolVersionsImpl address. So we mirror that value here so we can pass the intent validation.
	mustPutString(root, "superchainRoles.protocolVersionsOwner", stringWrapper("0x79ADD5713B383DAa0a138d3C4780C7A1804a8090"))

	return intent, nil
}

func StandardStateV4(l1ChainID uint64) (OpaqueState, error) {
	return standardState(l1ChainID, validation.Semver400, standardV4State)
}

func StandardStateV3(l1ChainID uint64) (OpaqueState, error) {
	return standardState(l1ChainID, validation.Semver300, standardV3State)
}

func StandardStateV2(l1ChainID uint64) (OpaqueState, error) {
	return standardState(l1ChainID, validation.Semver200, standardV2State)
}

func StandardStateV1(l1ChainID uint64) (OpaqueState, error) {
	return standardState(l1ChainID, validation.Semver180, standardV1State)
}

func standardState(l1ChainID uint64, semver validation.Semver, data []byte) (OpaqueState, error) {
	state := make(OpaqueState)
	if err := json.Unmarshal(data, &state); err != nil {
		panic(err)
	}

	var stdVersions validation.Versions
	var scNetwork string
	switch l1ChainID {
	case 1:
		stdVersions = validation.StandardVersionsMainnet
		scNetwork = "mainnet"
	case 11155111:
		stdVersions = validation.StandardVersionsSepolia
		scNetwork = "sepolia"
	default:
		return nil, fmt.Errorf("unsupported L1 chain ID: %d", l1ChainID)
	}

	stdVals, ok := stdVersions[semver]
	if !ok {
		return nil, fmt.Errorf("semver not found in stdVersions: %s", semver)
	}

	sc, err := superchain.GetSuperchain(scNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to get superchain: %w", err)
	}

	root := dasel.New(state)
	mustPutLowerString(root, "superchainDeployment.superchainConfigProxyAddress", sc.SuperchainConfigAddr)
	mustPutLowerString(root, "superchainDeployment.protocolVersionsProxyAddress", sc.ProtocolVersionsAddr)
	if stdVals.OPContractsManager != nil && stdVals.OPContractsManager.Address != nil {
		mustPutLowerString(root, "implementationsDeployment.opcmAddress", stdVals.OPContractsManager.Address)
	}
	mustPutLowerString(root, "implementationsDeployment.delayedWETHImplAddress", stdVals.DelayedWeth.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.optimismPortalImplAddress", stdVals.OptimismPortal.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.preimageOracleSingletonAddress", stdVals.PreimageOracle.Address)
	mustPutLowerString(root, "implementationsDeployment.mipsSingletonAddress", stdVals.Mips.Address)
	mustPutLowerString(root, "implementationsDeployment.systemConfigImplAddress", stdVals.SystemConfig.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.l1CrossDomainMessengerImplAddress", stdVals.L1CrossDomainMessenger.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.l1ERC721BridgeImplAddress", stdVals.L1ERC721Bridge.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.l1StandardBridgeImplAddress", stdVals.L1StandardBridge.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.optimismMintableERC20FactoryImplAddress", stdVals.OptimismMintableERC20Factory.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.disputeGameFactoryImplAddress", stdVals.DisputeGameFactory.ImplementationAddress)

	return state, nil
}

func mustPutString(node *dasel.Node, sel string, val fmt.Stringer) {
	if err := putString(node, sel, val); err != nil {
		panic(err)
	}
}

func mustPutLowerString(node *dasel.Node, sel string, val fmt.Stringer) {
	if err := node.Put(sel, strings.ToLower(val.String())); err != nil {
		panic(err)
	}
}

func putString(node *dasel.Node, sel string, val fmt.Stringer) error {
	return node.Put(sel, val.String())
}

func copyValue(src *dasel.Node, dest *dasel.Node, sel string) error {
	val, err := src.Query(sel)
	if err != nil {
		return fmt.Errorf("failed to read value: %w", err)
	}
	return dest.Put(sel, val.InterfaceValue())
}
