// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { console2 } from "forge-std/console2.sol";
import { Script } from "forge-std/Script.sol";
import { stdStorage } from "forge-std/StdStorage.sol";
import { StdAssertions } from "forge-std/StdAssertions.sol";

/**
 * @title CheckSecuityConfigs
 * @notice A script to check security configurations of an OP Chain,
           such as upgrade key holder, challenger and guadian designations.
 *         The usage is as follows:
 *         $ forge script CheckSecuityConfigs.s.sol \
 *             --rpc-url $ETH_RPC_URL
 */

contract CheckSecuityConfigs is Script, StdAssertions {

    struct ProtocolControllers {
        address FoundationMultisig;
        address CoinbaseMultisig;
        address ZoraMultisig;
        address PgnMultisig;
    }

    struct ProtocolContracts {
        // Please keep these sorted by name.
        address AddressManager;
        address L1CrossDomainMessengerProxy;
        address L1ERC721BridgeProxy;
        address L1StandardBridgeProxy;
        address L2OutputOracleProxy;
        address OptimismMintableERC20FactoryProxy;
        address OptimismPortalProxy;
        address ProxyAdmin;
    }

    /**
     * @notice The entrypoint function.
     */
    function run() external {
        string[] memory addressesJsonFiles = [vm.envOr("ADDRESSES_JSON")];
    }

    function runOnSingleFile(string memory addressesJsonPath) {
        ProtocolContracts memory contracts = getContracts(vm.readFile(addressesJsonPath));
        ProtocolContracts memory controllers = getControllers();
        checkAddressManager(contracts);
        checkL1CrossDomainMessengerProxy(contracts);
        checkL1ERC721BridgeProxy(contracts);
        checkL1StandardBridgeProxy(contracts);
        checkL2OutputOracleProxy(contracts);
        checkOptimismMintableERC20FactoryProxy(contracts);
        checkOptimismPortalProxy(contracts);
        // checkProxyAdmin(contracts);
    }

    function checkAddressManager(ContractSet memory contracts) internal {
        console2.log("Checking AddressManager %s", contracts.AddressManager);
        isOwnerOf(contracts.ProxyAdmin, contracts.AddressManager);
    }

    function checkL1CrossDomainMessengerProxy(ContractSet memory contracts) internal {
        console2.log("Checking L1CrossDomainMessengerProxy %s", contracts.L1CrossDomainMessengerProxy);

        address actualAddressManager = address(uint160(getMappingValue(contracts.L1CrossDomainMessengerProxy, 1, contracts.L1CrossDomainMessengerProxy)));
        assertEq(contracts.AddressManager, actualAddressManager);

        checkAddressIsExpected(contracts.OptimismPortalProxy, contracts.L1CrossDomainMessengerProxy, "PORTAL()");
    }

    function checkL1ERC721BridgeProxy(ContractSet memory contracts) internal {
        console2.log("Checking L1ERC721BridgeProxy %s", contracts.L1ERC721BridgeProxy);
        checkAddressIsExpected(contracts.L1ProxyAdmin, contracts.L1ERC721BridgeProxy, "admin()");
        checkAddressIsExpected(contracts.L1CrossDomainMessengerProxy, contracts.L1ERC721BridgeProxy, "messenger()");
    }

    function checkL1ProxyAdmin(ContractSet memory contracts) internal {
        console2.log("Checking L1ProxyAdmin %s", contracts.L1ProxyAdmin);
        checkAddressIsExpected(contracts.L1UpgradeKey, contracts.L1ProxyAdmin, "owner()");
    }

    function checkL1StandardBridgeProxy(ContractSet memory contracts) internal {
        console2.log("Checking L1StandardBridgeProxy %s", contracts.L1StandardBridgeProxy);
        checkAddressIsExpected(contracts.L1ProxyAdmin, contracts.L1StandardBridgeProxy, "getOwner()");
        checkAddressIsExpected(contracts.L1CrossDomainMessengerProxy, contracts.L1StandardBridgeProxy, "messenger()");
    }

    function checkL1UpgradeKey(ContractSet memory contracts) internal {
        console2.log("Checking L1UpgradeKeyAddress %s", contracts.L1UpgradeKey);
        // No need to check anything here, so just printing the address.
    }

    function checkL2OutputOracleProxy(ContractSet memory contracts) internal {
        console2.log("Checking L2OutputOracleProxy %s", contracts.L2OutputOracleProxy);
        checkAddressIsExpected(contracts.L1ProxyAdmin, contracts.L2OutputOracleProxy, "admin()");
        checkAddressIsExpected(contracts.L1ChallengerKey, contracts.L2OutputOracleProxy, "CHALLENGER()");
        // 604800 seconds = 7 days, reusing the logic in
        // checkAddressIsExpected for simplicity.
        checkAddressIsExpected(address(604800), contracts.L2OutputOracleProxy, "FINALIZATION_PERIOD_SECONDS()");
    }

    function checkOptimismMintableERC20FactoryProxy(ContractSet memory contracts) internal {
        console2.log("Checking OptimismMintableERC20FactoryProxy %s", contracts.OptimismMintableERC20FactoryProxy);
        checkAddressIsExpected(contracts.L1ProxyAdmin, contracts.OptimismMintableERC20FactoryProxy, "admin()");
        checkAddressIsExpected(contracts.L1StandardBridgeProxy, contracts.OptimismMintableERC20FactoryProxy, "BRIDGE()");
    }

    function checkOptimismPortalProxy(ContractSet memory contracts) internal {
        console2.log("Checking OptimismPortalProxy %s", contracts.OptimismPortalProxy);
        checkAddressIsExpected(contracts.L1ProxyAdmin, contracts.OptimismPortalProxy, "admin()");
        checkAddressIsExpected(contracts.L1ChallengerKey, contracts.OptimismPortalProxy, "GUARDIAN()");
        checkAddressIsExpected(contracts.L2OutputOracleProxy, contracts.OptimismPortalProxy, "L2_ORACLE()");
    }

    /* function checkSystemConfigProxy(ContractSet memory contracts) internal { */
    /*     console2.log("Checking SystemConfigProxy %s", contracts.SystemConfigProxy); */
    /*     checkAddressIsExpected(contracts.L1ProxyAdmin, contracts.SystemConfigProxy, "admin()"); */
    /* } */


    function getMappingValue(address targetContract, uint256 mapSlot, address key) public view returns (uint256) {
        bytes32 slotValue = vm.load(targetContract, keccak256(abi.encode(key, mapSlot)));
        return uint256(slotValue);
    }

    function uint2str(uint256 from) internal pure returns (string memory str) {
        bytes memory bytesString = new bytes(32);
        for (uint j=0; j<32; j++) {
            bytesString[32-j-1] = bytes1(uint8(from));
            from = from / 0x100;
        }
        bool gotZero = false;
        for (uint j=0; j<32; j++) {
            if (gotZero) {
                bytesString[j] = 0;
            } else {
                gotZero = bytesString[j] == 0;
            }
        }
        bytesString[31] = 0;
        return string(bytesString);
    }

    function isOwnerOf(address expectedOwner, address ownableContract) internal {
        checkAddressIsExpected(expectedOwner, ownableContract, "owner()");
    }

    function checkAddressIsExpected(address expectedAddr, address contractAddr, string memory signature) internal {
        address actual = getAddressFromCall(contractAddr, signature);
        if (expectedAddr != actual) {
            console2.log("  !! Error: %s != %s.%s, ", expectedAddr, contractAddr, signature);
            console2.log("           which is %s", actual);
        } else {
            console2.log("  -- Success: %s == %s.%s.", expectedAddr, contractAddr, signature);
        }
    }

    function getAddressFromCall(address contractAddr, string memory signature) internal returns (address) {
        vm.prank(address(0));
        (bool success, bytes memory addrBytes) = contractAddr.staticcall(abi.encodeWithSignature(signature));
        if (!success) {
            console2.log("  !! Error calling %s.%s", contractAddr, signature);
            return address(0);
        }
        return abi.decode(addrBytes, (address));
    }

    function getContracts(string memory addressesJson) internal returns (ProtocolContracts memory) {
        return ProtocolContracts({
            AddressManager: vm.parseJsonAddress(addressesJson, ".AddressManager"),
            L1CrossDomainMessengerProxy: vm.parseJsonAddress(addressesJson, ".L1CrossDomainMessengerProxy"),
            L1ERC721BridgeProxy: vm.parseJsonAddress(addressesJson, ".L1ERC721BridgeProxy"),
            L1StandardBridgeProxy: vm.parseJsonAddress(addressesJson, ".L1StandardBridgeProxy"),
            L2OutputOracleProxy: vm.parseJsonAddress(addressesJson, ".L2OutputOracleProxy"),
            OptimismMintableERC20FactoryProxy: vm.parseJsonAddress(addressesJson, ".OptimismMintableERC20FactoryProxy"),
            OptimismPortalProxy: vm.parseJsonAddress(addressesJson, ".OptimismPortalProxy"),
            ProxyAdmin: vm.parseJsonAddress(addressesJson, ".ProxyAdmin"),
            });
    }

    function getControllers() internal returns (ProtocolControllers memory) {
        return ProtocolControllers({
            FoundationMultisig: 0x9BA6e03D8B90dE867373Db8cF1A58d2F7F006b3A,
            CoinbaseMultisig: 0x9855054731540A48b28990B63DcF4f33d8AE46A1,
            ZoraMultisig: 0xC72aE5c7cc9a332699305E29F68Be66c73b60542,
            PgnMultisig: 0x4a4962275DF8C60a80d3a25faEc5AA7De116A746
            });
    }
}
