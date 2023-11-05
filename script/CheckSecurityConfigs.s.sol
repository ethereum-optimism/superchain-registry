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

    struct ContractSet {
        // Please keep these sorted by name.
        address AddressManager;
        address L1CrossDomainMessengerImpl;
        address L1CrossDomainMessengerProxy;
        address L1ERC721BridgeImpl;
        address L1ERC721BridgeProxy;
        address L1ProxyAdmin;
        address L1StandardBridgeImpl;
        address L1StandardBridgeProxy;
        address L1ChallengerKey;
        address L1UpgradeKey;
        address L2OutputOracleImpl;
        address L2OutputOracleProxy;
        address OptimismMintableERC20FactoryImpl;
        address OptimismMintableERC20FactoryProxy;
        address OptimismPortalImpl;
        address OptimismPortalProxy;
        address SystemConfigProxy;
    }

    /**
     * @notice The entrypoint function.
     */
    function run() external {
        string memory bedrockJsonDir = vm.envString("BEDROCK_JSON_DIR"); // deployments/zora;
        console2.log("BEDROCK_JSON_DIR = %s", bedrockJsonDir);
        console2.log("L1_UPGRADE_KEY = %s", vm.envAddress("L1_UPGRADE_KEY"));
        console2.log("L1_CHALLENGER_KEY = %s", vm.envAddress("L1_CHALLENGER_KEY"));
        ContractSet memory contracts = getContracts(bedrockJsonDir);
        checkAddressManager(contracts);
        checkL1CrossDomainMessengerProxy(contracts);
        checkL1ERC721BridgeProxy(contracts);
        checkL1ProxyAdmin(contracts);
        checkL1StandardBridgeProxy(contracts);
        checkL1UpgradeKey(contracts);
        checkL2OutputOracleProxy(contracts);
        checkOptimismMintableERC20FactoryProxy(contracts);
        checkOptimismPortalProxy(contracts);
        checkSystemConfigProxy(contracts);
    }

    function checkAddressManager(ContractSet memory contracts) internal {
        console2.log("Checking AddressManager %s", contracts.AddressManager);
        checkAddressIsExpected(contracts.L1ProxyAdmin, contracts.AddressManager, "owner()");
    }

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

    function checkL1CrossDomainMessengerProxy(ContractSet memory contracts) internal {
        console2.log("Checking L1CrossDomainMessengerProxy %s", contracts.L1CrossDomainMessengerProxy);

        address addressManager = address(uint160(getMappingValue(contracts.L1CrossDomainMessengerProxy, 1, contracts.L1CrossDomainMessengerProxy)));
        checkAddressIsExpected(contracts.L1ProxyAdmin, addressManager, "owner()");

        string memory implementationName = uint2str(getMappingValue(contracts.L1CrossDomainMessengerProxy, 0, contracts.L1CrossDomainMessengerProxy));
        console2.log("  !!! TODO: check %s == cast call --flashbots %s \"getAddress(string)(address)\" \"%s\"", contracts.L1CrossDomainMessengerImpl, addressManager, implementationName);
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

    function checkSystemConfigProxy(ContractSet memory contracts) internal {
        console2.log("Checking SystemConfigProxy %s", contracts.SystemConfigProxy);
        checkAddressIsExpected(contracts.L1ProxyAdmin, contracts.SystemConfigProxy, "admin()");
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

    function getContracts(string memory bedrockJsonDir) internal returns (ContractSet memory) {
        return ContractSet({
                AddressManager: getAddressFromJson(string.concat(bedrockJsonDir, "/Lib_AddressManager.json")),
                L1CrossDomainMessengerImpl: getAddressFromJson(string.concat(bedrockJsonDir, "/L1CrossDomainMessenger.json")),
                L1CrossDomainMessengerProxy: getAddressFromJson(string.concat(bedrockJsonDir, "/Proxy__OVM_L1CrossDomainMessenger.json")),
                L1ERC721BridgeImpl: getAddressFromJson(string.concat(bedrockJsonDir, "/L1ERC721Bridge.json")),
                L1ERC721BridgeProxy: getAddressFromJson(string.concat(bedrockJsonDir, "/L1ERC721BridgeProxy.json")),
                L1ProxyAdmin: getAddressFromJson(string.concat(bedrockJsonDir, "/ProxyAdmin.json")),
                L1StandardBridgeImpl: getAddressFromJson(string.concat(bedrockJsonDir, "/L1StandardBridge.json")),
                L1StandardBridgeProxy: getAddressFromJson(string.concat(bedrockJsonDir, "/Proxy__OVM_L1StandardBridge.json")),
                L1ChallengerKey: vm.envAddress("L1_CHALLENGER_KEY"), //0xcA4571b1ecBeC86Ea2E660d242c1c29FcB55Dc72,
                L1UpgradeKey: vm.envAddress("L1_UPGRADE_KEY"), //0xC72aE5c7cc9a332699305E29F68Be66c73b60542,
                L2OutputOracleImpl: getAddressFromJson(string.concat(bedrockJsonDir, "/L2OutputOracle.json")),
                L2OutputOracleProxy: getAddressFromJson(string.concat(bedrockJsonDir, "/L2OutputOracleProxy.json")),
                OptimismMintableERC20FactoryImpl: getAddressFromJson(string.concat(bedrockJsonDir, "/OptimismMintableERC20Factory.json")),
                OptimismMintableERC20FactoryProxy: getAddressFromJson(string.concat(bedrockJsonDir, "/OptimismMintableERC20FactoryProxy.json")),
                OptimismPortalImpl: getAddressFromJson(string.concat(bedrockJsonDir, "/OptimismPortal.json")),
                OptimismPortalProxy: getAddressFromJson(string.concat(bedrockJsonDir, "/OptimismPortalProxy.json")),
                SystemConfigProxy: getAddressFromJson(string.concat(bedrockJsonDir, "/SystemConfigProxy.json"))
            });
    }

    function getAddressFromJson(string memory jsonPath) internal returns (address) {
        string memory json = vm.readFile(jsonPath);
        return vm.parseJsonAddress(json, ".address");
    }

}
