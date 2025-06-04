package state

import (
	_ "embed"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer/opaque_map"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/hashicorp/go-multierror"
	"github.com/tomwright/dasel"
)

//go:embed configs/v2-state.json
var standardV2State []byte

//go:embed configs/v2-intent.toml
var standardV2Intent []byte

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

func StandardStateV2(l1ChainID uint64) (OpaqueState, error) {
	return standardState(l1ChainID, validation.Semver200, standardV2State)
}

func StandardIntentV2(l1ChainID uint64) (opaque_map.OpaqueMap, error) {
	return standardIntentV2(l1ChainID, standardV2Intent)
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
