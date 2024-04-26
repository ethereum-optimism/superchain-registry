// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import {console2} from "forge-std/console2.sol";
import {Script} from "forge-std/Script.sol";
import {VmSafe} from "forge-std/Vm.sol";

/**
 * @title CheckSecurityConfigs
 * @notice A script to check security configurations of an OP Chain,
 *            such as upgrade key holder, challenger and guadian designations.
 *         The usage is as follows:
 *         $ forge script CheckSecurityConfigs \
 *             --rpc-url $MAINNET_RPC_URL
 */
contract CheckSecurityConfigs is Script {
    struct ProtocolAddresses {
        // Protocol contracts
        address AddressManager;
        address L1CrossDomainMessengerProxy;
        address L1ERC721BridgeProxy;
        address L1StandardBridgeProxy;
        address OptimismMintableERC20FactoryProxy;
        address OptimismPortalProxy;
        address ProxyAdmin;
        address SystemConfigProxy;
        // Legacy pre-fault proof protocol contracts
        address L2OutputOracleProxy;
        // Fault Proof protocol contracts
        address DisputeGameFactoryProxy;
        address AnchorStateRegistryProxy;
        address DelayedWETHProxy;
        // Privileged roles
        address Challenger;
        address Guardian;
        address ProxyAdminOwner;
        address SystemConfigOwner;
    }

    bool internal hasErrors;

    /**
     * @notice The entrypoint function.
     */
    function run() public {
        string memory network;
        if (block.chainid == 1) {
            network = "mainnet";
        } else if (block.chainid == 11155111) {
            network = "sepolia";
        } else if (block.chainid == 5) {
            network = "goerli";
        } else {
            revert(
                string.concat(
                    "Unsupported chain ID: ", vm.toString(block.chainid), ". Please call runOnDir(string) directly."
                )
            );
        }
        string memory jsonDir = string.concat("superchain/extra/addresses/", network);
        runOnDir(jsonDir, block.chainid == 1);
    }

    function runOnDir(string memory jsonDir, bool isMainnet) public {
        VmSafe.DirEntry[] memory addressesJsonEntries = vm.readDir(jsonDir);
        hasErrors = false;
        for (uint256 i = 0; i < addressesJsonEntries.length; i++) {
            require(bytes(addressesJsonEntries[i].errorMessage).length == 0, addressesJsonEntries[i].errorMessage);
            runOnSingleFile(addressesJsonEntries[i].path, isMainnet);
        }
        require(!hasErrors, "Errors occurred: See logs above for more info");
    }

    function runOnSingleFile(string memory addressesJsonPath, bool isMainnet) internal {
        console2.log("Checking %s", addressesJsonPath);

        (ProtocolAddresses memory addresses, bool upgradedToFPAC) = getAddressesAndFpacStatus(addressesJsonPath);
        checkAddressManager(addresses);
        checkL1CrossDomainMessengerProxy(addresses);
        checkL1ERC721BridgeProxy(addresses);
        checkL1StandardBridgeProxy(addresses);
        checkOptimismMintableERC20FactoryProxy(addresses);
        checkOptimismPortalProxy(addresses, upgradedToFPAC);
        checkProxyAdmin(addresses);
        checkSystemConfigProxy(addresses);
        if (upgradedToFPAC) {
            checkDisputeGameFactoryProxy(addresses);
            checkAnchorStateRegistryProxy(addresses);
            checkDelayedWETHProxy(addresses);
        } else {
            checkL2OutputOracleProxy(addresses, isMainnet, upgradedToFPAC);
        }
        // TODO Check the integrity of the implementations: https://github.com/ethereum-optimism/superchain-registry/issues/33
    }

    function chainUpgradedToFPAC(ProtocolAddresses memory addresses) internal view returns (bool) {
        ISemver portalProxy = ISemver(addresses.OptimismPortalProxy);
        string memory version = portalProxy.version();
        string memory majorVersion = sliceString(version, 0, 1);

        // Portal version `3` is the first version of the `OptimismPortal` that supported the fault proof system.
        return vm.parseUint(majorVersion) >= 3;
    }

    // Function to slice a string and return the result
    function sliceString(string memory str, uint256 begin, uint256 end) public pure returns (string memory) {
        require(begin < end, "Begin index must be less than end index");
        bytes memory strBytes = bytes(str);
        require(end <= strBytes.length, "End index out of bounds");

        bytes memory result = new bytes(end - begin);
        for (uint256 i = begin; i < end; i++) {
            result[i - begin] = strBytes[i];
        }

        return string(result);
    }

    function checkAddressManager(ProtocolAddresses memory addresses) internal {
        console2.log("Checking AddressManager %s", addresses.AddressManager);
        isOwnerOf(addresses.ProxyAdmin, addresses.AddressManager);
    }

    function checkL1CrossDomainMessengerProxy(ProtocolAddresses memory addresses) internal {
        console2.log("Checking L1CrossDomainMessengerProxy %s", addresses.L1CrossDomainMessengerProxy);

        address actualAddressManager = address(
            uint160(getMappingValue(addresses.L1CrossDomainMessengerProxy, 1, addresses.L1CrossDomainMessengerProxy))
        );
        assert(addresses.AddressManager == actualAddressManager);

        checkAddressIsExpected(addresses.OptimismPortalProxy, addresses.L1CrossDomainMessengerProxy, "PORTAL()");
    }

    function checkL1ERC721BridgeProxy(ProtocolAddresses memory addresses) internal {
        console2.log("Checking L1ERC721BridgeProxy %s", addresses.L1ERC721BridgeProxy);
        isAdminOf(addresses.ProxyAdmin, addresses.L1ERC721BridgeProxy);
        checkAddressIsExpected(addresses.L1CrossDomainMessengerProxy, addresses.L1ERC721BridgeProxy, "messenger()");
    }

    function checkL1StandardBridgeProxy(ProtocolAddresses memory addresses) internal {
        console2.log("Checking L1StandardBridgeProxy %s", addresses.L1StandardBridgeProxy);
        checkAddressIsExpected(addresses.ProxyAdmin, addresses.L1StandardBridgeProxy, "getOwner()");
        checkAddressIsExpected(addresses.L1CrossDomainMessengerProxy, addresses.L1StandardBridgeProxy, "messenger()");
    }

    function checkL2OutputOracleProxy(ProtocolAddresses memory addresses, bool isMainnet, bool upgradedToFPAC)
        internal
    {
        if (upgradedToFPAC) {
            // This check is skipped for chains which upgraded to FPAC
            console2.log("Skipping L2OutputOracleProxy check for FPAC enabled chain");
            return;
        }
        console2.log("Checking L2OutputOracleProxy %s", addresses.L2OutputOracleProxy);
        isAdminOf(addresses.ProxyAdmin, addresses.L2OutputOracleProxy);
        checkAddressIsExpected(addresses.Challenger, addresses.L2OutputOracleProxy, "CHALLENGER()");
        if (isMainnet) {
            // Reusing the logic in checkAddressIsExpected below for simplicity.
            checkAddressIsExpected(address(7 days), addresses.L2OutputOracleProxy, "FINALIZATION_PERIOD_SECONDS()");
        }
    }

    function checkOptimismMintableERC20FactoryProxy(ProtocolAddresses memory addresses) internal {
        console2.log("Checking OptimismMintableERC20FactoryProxy %s", addresses.OptimismMintableERC20FactoryProxy);
        isAdminOf(addresses.ProxyAdmin, addresses.OptimismMintableERC20FactoryProxy);
        checkAddressIsExpected(addresses.L1StandardBridgeProxy, addresses.OptimismMintableERC20FactoryProxy, "BRIDGE()");
    }

    function checkOptimismPortalProxy(ProtocolAddresses memory addresses, bool upgradedToFPAC) internal {
        // Legacy immutable `SCREAMING_SNAKE_CASE` function signatures were deprecated in FPAC.
        string[] memory fnSigs = new string[](2);
        if (upgradedToFPAC) {
            fnSigs[0] = "guardian()";
            fnSigs[1] = "systemConfig()";
        } else {
            fnSigs[0] = "GUARDIAN()";
            fnSigs[1] = "SYSTEM_CONFIG()";
        }

        console2.log("Checking OptimismPortalProxy %s", addresses.OptimismPortalProxy);
        isAdminOf(addresses.ProxyAdmin, addresses.OptimismPortalProxy);
        checkAddressIsExpected(addresses.Guardian, addresses.OptimismPortalProxy, fnSigs[0]);
        if (!upgradedToFPAC) {
            checkAddressIsExpected(addresses.L2OutputOracleProxy, addresses.OptimismPortalProxy, "L2_ORACLE()");
        }
        checkAddressIsExpected(addresses.SystemConfigProxy, addresses.OptimismPortalProxy, fnSigs[1]);
    }

    function checkSystemConfigProxy(ProtocolAddresses memory addresses) internal {
        console2.log("Checking SystemConfigProxy %s", addresses.SystemConfigProxy);
        isAdminOf(addresses.ProxyAdmin, addresses.SystemConfigProxy);
        isOwnerOf(addresses.SystemConfigOwner, addresses.SystemConfigProxy);
    }

    function checkDisputeGameFactoryProxy(ProtocolAddresses memory addresses) internal {
        console2.log("Checking DisputeGameFactoryProxy %s", addresses.DisputeGameFactoryProxy);
        isAdminOf(addresses.ProxyAdmin, addresses.DisputeGameFactoryProxy);
        isOwnerOf(addresses.ProxyAdminOwner, addresses.DisputeGameFactoryProxy);
    }

    function checkAnchorStateRegistryProxy(ProtocolAddresses memory addresses) internal {
        console2.log("Checking AnchorStateRegistryProxy %s", addresses.AnchorStateRegistryProxy);
        isAdminOf(addresses.ProxyAdmin, addresses.AnchorStateRegistryProxy);
    }

    function checkDelayedWETHProxy(ProtocolAddresses memory addresses) internal {
        console2.log("Checking DelayedWETHProxy %s", addresses.DelayedWETHProxy);
        isAdminOf(addresses.ProxyAdmin, addresses.DelayedWETHProxy);
        isOwnerOf(addresses.ProxyAdminOwner, addresses.DelayedWETHProxy);
    }

    function checkProxyAdmin(ProtocolAddresses memory addresses) internal {
        console2.log("Checking ProxyAdmin %s", addresses.ProxyAdmin);
        isOwnerOf(addresses.ProxyAdminOwner, addresses.ProxyAdmin);
        checkAddressIsExpected(addresses.AddressManager, addresses.ProxyAdmin, "addressManager()");
    }

    function getMappingValue(address targetContract, uint256 mapSlot, address key) public view returns (uint256) {
        bytes32 slotValue = vm.load(targetContract, keccak256(abi.encode(key, mapSlot)));
        return uint256(slotValue);
    }

    function isAdminOf(address expectedOwner, address ownableContract) internal {
        checkAddressIsExpected(expectedOwner, ownableContract, "admin()");
    }

    function isOwnerOf(address expectedOwner, address ownableContract) internal {
        checkAddressIsExpected(expectedOwner, ownableContract, "owner()");
    }

    function checkAddressIsExpected(address expectedAddr, address contractAddr, string memory signature) internal {
        address actual = getAddressFromCall(contractAddr, signature);
        if (expectedAddr != actual) {
            console2.log("  !! Error: %s != %s.%s, ", expectedAddr, contractAddr, signature);
            console2.log("           which is %s", actual);
            hasErrors = true;
        } else {
            console2.log("  -- Success: %s == %s.%s.", expectedAddr, contractAddr, signature);
        }
    }

    function getAddressFromCall(address contractAddr, string memory signature) internal returns (address) {
        vm.prank(address(0));
        (bool success, bytes memory addrBytes) = contractAddr.staticcall(abi.encodeWithSignature(signature));
        if (!success) {
            console2.log("  !! Error calling %s.%s", contractAddr, signature);
            hasErrors = true;
            return address(0);
        }
        return abi.decode(addrBytes, (address));
    }

    function getAddressesAndFpacStatus(string memory addressesJsonPath)
        internal
        view
        returns (ProtocolAddresses memory, bool)
    {
        string memory addressesJson = vm.readFile(addressesJsonPath);
        ProtocolAddresses memory addresses = ProtocolAddresses({
            AddressManager: vm.parseJsonAddress(addressesJson, ".AddressManager"),
            L1CrossDomainMessengerProxy: vm.parseJsonAddress(addressesJson, ".L1CrossDomainMessengerProxy"),
            L1ERC721BridgeProxy: vm.parseJsonAddress(addressesJson, ".L1ERC721BridgeProxy"),
            L1StandardBridgeProxy: vm.parseJsonAddress(addressesJson, ".L1StandardBridgeProxy"),
            OptimismMintableERC20FactoryProxy: vm.parseJsonAddress(addressesJson, ".OptimismMintableERC20FactoryProxy"),
            OptimismPortalProxy: vm.parseJsonAddress(addressesJson, ".OptimismPortalProxy"),
            ProxyAdmin: vm.parseJsonAddress(addressesJson, ".ProxyAdmin"),
            SystemConfigProxy: vm.parseJsonAddress(addressesJson, ".SystemConfigProxy"),
            Challenger: vm.parseJsonAddress(addressesJson, ".Challenger"),
            Guardian: vm.parseJsonAddress(addressesJson, ".Guardian"),
            ProxyAdminOwner: vm.parseJsonAddress(addressesJson, ".ProxyAdminOwner"),
            SystemConfigOwner: vm.parseJsonAddress(addressesJson, ".SystemConfigOwner"),
            // Unset initially; dependent on whether the chain supports FPAC.
            L2OutputOracleProxy: address(0),
            DisputeGameFactoryProxy: address(0),
            AnchorStateRegistryProxy: address(0),
            DelayedWETHProxy: address(0)
        });

        bool upgradedToFPAC = chainUpgradedToFPAC(addresses);
        if (upgradedToFPAC) {
            addresses.DisputeGameFactoryProxy = vm.parseJsonAddress(addressesJson, ".DisputeGameFactoryProxy");
            addresses.AnchorStateRegistryProxy = vm.parseJsonAddress(addressesJson, ".AnchorStateRegistryProxy");
            addresses.DelayedWETHProxy = vm.parseJsonAddress(addressesJson, ".DelayedWETHProxy");
        } else {
            addresses.L2OutputOracleProxy = vm.parseJsonAddress(addressesJson, ".L2OutputOracleProxy");
        }

        return (addresses, upgradedToFPAC);
    }
}

interface ISemver {
    function version() external view returns (string memory version);
}
