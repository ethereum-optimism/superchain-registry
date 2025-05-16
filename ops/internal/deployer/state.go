package deployer

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/superchain"
	"github.com/hashicorp/go-multierror"
	"github.com/tomwright/dasel"
)

//go:embed configs/v2-state.json
var standardV2State []byte

//go:embed configs/v2-intent.toml
var standardV2Intent []byte

//go:embed configs/v3-state.json
var standardV3State []byte

//go:embed configs/v3-intent.toml
var standardV3Intent []byte

type OpaqueMapping map[string]any

func ReadOpaqueMappingFile(p string) (OpaqueMapping, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer f.Close()

	var out OpaqueMapping
	if err := json.NewDecoder(f).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return out, nil
}

func MergeStateV2(userState OpaqueMapping) (OpaqueMapping, OpaqueMapping, error) {
	l1ChainID, err := readL1ChainID(dasel.New(userState))
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

func MergeStateV3(userState OpaqueMapping) (OpaqueMapping, OpaqueMapping, error) {
	l1ChainID, err := readL1ChainID(dasel.New(userState))
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
	return mergeStateV2(userState, stdIntent, stdState)
}

func mergeStateV2(userState OpaqueMapping, stdIntent OpaqueMapping, stdState OpaqueMapping) (OpaqueMapping, OpaqueMapping, error) {
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

	if copyErrs != nil {
		return nil, nil, copyErrs
	}

	intentResult, okIntent := stdIntentNode.InterfaceValue().(OpaqueMapping)
	if !okIntent {
		return nil, nil, fmt.Errorf("internal error: synthesized intent is not OpaqueMapping, but %T", stdIntentNode.InterfaceValue())
	}
	stateResult, okState := stdStateNode.InterfaceValue().(OpaqueMapping)
	if !okState {
		return nil, nil, fmt.Errorf("internal error: synthesized state is not OpaqueMapping, but %T", stdStateNode.InterfaceValue())
	}

	return intentResult, stateResult, nil
}

func StandardIntentV3(l1ChainID uint64) (OpaqueMapping, error) {
	return standardIntentV2(l1ChainID, standardV3Intent)
}

func StandardIntentV2(l1ChainID uint64) (OpaqueMapping, error) {
	return standardIntentV2(l1ChainID, standardV2Intent)
}

func standardIntentV2(l1ChainID uint64, data []byte) (OpaqueMapping, error) {
	intent := make(OpaqueMapping)
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

func StandardStateV3(l1ChainID uint64) (OpaqueMapping, error) {
	return standardState(l1ChainID, validation.Semver300, standardV3State)
}

func StandardStateV2(l1ChainID uint64) (OpaqueMapping, error) {
	return standardState(l1ChainID, validation.Semver200, standardV2State)
}

func standardState(l1ChainID uint64, semver validation.Semver, data []byte) (OpaqueMapping, error) {
	state := make(OpaqueMapping)
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

	v2Info := stdVersions[semver]

	sc, err := superchain.GetSuperchain(scNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to get superchain: %w", err)
	}

	root := dasel.New(state)
	mustPutLowerString(root, "superchainDeployment.superchainConfigProxyAddress", sc.SuperchainConfigAddr)
	mustPutLowerString(root, "superchainDeployment.protocolVersionsProxyAddress", sc.ProtocolVersionsAddr)
	mustPutLowerString(root, "implementationsDeployment.opcmAddress", v2Info.OPContractsManager.Address)
	mustPutLowerString(root, "implementationsDeployment.delayedWETHImplAddress", v2Info.DelayedWeth.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.optimismPortalImplAddress", v2Info.OptimismPortal.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.preimageOracleSingletonAddress", v2Info.PreimageOracle.Address)
	mustPutLowerString(root, "implementationsDeployment.mipsSingletonAddress", v2Info.Mips.Address)
	mustPutLowerString(root, "implementationsDeployment.systemConfigImplAddress", v2Info.SystemConfig.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.l1CrossDomainMessengerImplAddress", v2Info.L1CrossDomainMessenger.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.l1ERC721BridgeImplAddress", v2Info.L1ERC721Bridge.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.l1StandardBridgeImplAddress", v2Info.L1StandardBridge.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.optimismMintableERC20FactoryImplAddress", v2Info.OptimismMintableERC20Factory.ImplementationAddress)
	mustPutLowerString(root, "implementationsDeployment.disputeGameFactoryImplAddress", v2Info.DisputeGameFactory.ImplementationAddress)

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

func readL1ChainID(node *dasel.Node) (uint64, error) {
	l1ChainIDNode, err := node.Query("appliedIntent.l1ChainID")
	if err != nil {
		return 0, fmt.Errorf("failed to read L1 chain ID: %w", err)
	}
	l1ChainIDFloat, ok := l1ChainIDNode.InterfaceValue().(float64)
	if !ok {
		return 0, errors.New("failed to parse L1 chain ID")
	}
	return uint64(l1ChainIDFloat), nil
}
