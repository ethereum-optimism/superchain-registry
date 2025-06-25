package state

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer/opaque_map"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/superchain"
	"github.com/tomwright/dasel"
)

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

// Add this type near the top of the file
type stringWrapper string

func (s stringWrapper) String() string {
	return string(s)
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
