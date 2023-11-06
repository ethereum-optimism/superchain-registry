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
 *         $ forge script CheckSecuityConfigs \
 *             --rpc-url $ETH_RPC_URL
 */

contract CheckSecuityConfigs is Script, StdAssertions {
    struct ProtocolControllers {
        address FoundationMultisig;

        address BaseChallenger1of2;
        address BaseOpsMultisig;
        address BaseUpgradeMultisig;

        address PgnOpsMultisig;
        address PgnUpgradeMultisig;

        address ZoraChallengerMultisig;
        address ZoraGuardianMultisig;
        address ZoraUpgradeMultisig;
    }
    ProtocolControllers controllers;

    struct ProtocolContracts {
        // ProtocolContracts
        address AddressManager;
        address L1CrossDomainMessengerProxy;
        address L1ERC721BridgeProxy;
        address L1StandardBridgeProxy;
        address L2OutputOracleProxy;
        address OptimismMintableERC20FactoryProxy;
        address OptimismPortalProxy;
        address ProxyAdmin;

        // Roles
        address ProxyAdminOwner;
        address Challenger;
        address Guardian;
    }

    mapping(string => address) proxyAdminOwnerExceptions;
    mapping(string => address) challengerExceptions;
    mapping(string => address) guardianExceptions;

    /**
     * @notice The entrypoint function.
     */
    function run() external {
        initializeControllers();
        initializeExceptions();
        string[4] memory addressesJsonFiles = [
            "superchain/extra/addresses/mainnet/base.json",
            "superchain/extra/addresses/mainnet/op.json",
            "superchain/extra/addresses/mainnet/pgn.json",
            "superchain/extra/addresses/mainnet/zora.json"
        ];
        for(uint i = 0; i < addressesJsonFiles.length; i++) {
            runOnSingleFile(addressesJsonFiles[i]);
        }
    }

    function runOnSingleFile(string memory addressesJsonPath) internal {
        console2.log("Checking %s", addressesJsonPath);
        ProtocolContracts memory contracts = getContracts(addressesJsonPath);
        checkAddressManager(contracts);
        checkL1CrossDomainMessengerProxy(contracts);
        checkL1ERC721BridgeProxy(contracts);
        checkL1StandardBridgeProxy(contracts);
        checkL2OutputOracleProxy(contracts);
        checkOptimismMintableERC20FactoryProxy(contracts);
        checkOptimismPortalProxy(contracts);
        checkProxyAdmin(contracts);
    }

    function checkAddressManager(ProtocolContracts memory contracts) internal {
        console2.log("Checking AddressManager %s", contracts.AddressManager);
        isOwnerOf(contracts.ProxyAdmin, contracts.AddressManager);
    }

    function checkL1CrossDomainMessengerProxy(ProtocolContracts memory contracts) internal {
        console2.log("Checking L1CrossDomainMessengerProxy %s", contracts.L1CrossDomainMessengerProxy);

        address actualAddressManager = address(uint160(getMappingValue(contracts.L1CrossDomainMessengerProxy, 1, contracts.L1CrossDomainMessengerProxy)));
        assertEq(contracts.AddressManager, actualAddressManager);

        checkAddressIsExpected(contracts.OptimismPortalProxy, contracts.L1CrossDomainMessengerProxy, "PORTAL()");
    }

    function checkL1ERC721BridgeProxy(ProtocolContracts memory contracts) internal {
        console2.log("Checking L1ERC721BridgeProxy %s", contracts.L1ERC721BridgeProxy);
        isAdminOf(contracts.ProxyAdmin, contracts.L1ERC721BridgeProxy);
        checkAddressIsExpected(contracts.L1CrossDomainMessengerProxy, contracts.L1ERC721BridgeProxy, "messenger()");
    }

    function checkL1StandardBridgeProxy(ProtocolContracts memory contracts) internal {
        console2.log("Checking L1StandardBridgeProxy %s", contracts.L1StandardBridgeProxy);
        checkAddressIsExpected(contracts.ProxyAdmin, contracts.L1StandardBridgeProxy, "getOwner()");
        checkAddressIsExpected(contracts.L1CrossDomainMessengerProxy, contracts.L1StandardBridgeProxy, "messenger()");
        checkAddressIsExpected(contracts.L1CrossDomainMessengerProxy, contracts.L1StandardBridgeProxy, "MESSENGER()");
    }

    function checkL2OutputOracleProxy(ProtocolContracts memory contracts) internal {
        console2.log("Checking L2OutputOracleProxy %s", contracts.L2OutputOracleProxy);
        isAdminOf(contracts.ProxyAdmin, contracts.L2OutputOracleProxy);
        checkAddressIsExpected(contracts.Challenger, contracts.L2OutputOracleProxy, "CHALLENGER()");
        // 604800 seconds = 7 days, reusing the logic in
        // checkAddressIsExpected for simplicity.
        checkAddressIsExpected(address(604800), contracts.L2OutputOracleProxy, "FINALIZATION_PERIOD_SECONDS()");
    }

    function checkOptimismMintableERC20FactoryProxy(ProtocolContracts memory contracts) internal {
        console2.log("Checking OptimismMintableERC20FactoryProxy %s", contracts.OptimismMintableERC20FactoryProxy);
        isAdminOf(contracts.ProxyAdmin, contracts.OptimismMintableERC20FactoryProxy);
        checkAddressIsExpected(contracts.L1StandardBridgeProxy, contracts.OptimismMintableERC20FactoryProxy, "BRIDGE()");
    }

    function checkOptimismPortalProxy(ProtocolContracts memory contracts) internal {
        console2.log("Checking OptimismPortalProxy %s", contracts.OptimismPortalProxy);
        isAdminOf(contracts.ProxyAdmin, contracts.OptimismPortalProxy);
        checkAddressIsExpected(contracts.Guardian, contracts.OptimismPortalProxy, "GUARDIAN()");
        checkAddressIsExpected(contracts.L2OutputOracleProxy, contracts.OptimismPortalProxy, "L2_ORACLE()");
        // TODO: Check SYSTEM_CONFIG()
    }

    // TODO: implement checkSystemConfigProxy(...)

    function checkProxyAdmin(ProtocolContracts memory contracts) internal {
        console2.log("Checking ProxyAdmin %s", contracts.ProxyAdmin);
        isOwnerOf(contracts.ProxyAdminOwner, contracts.ProxyAdmin);
        checkAddressIsExpected(contracts.AddressManager, contracts.ProxyAdmin, "addressManager()");
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

    function getContracts(string memory addressesJsonPath) internal view returns (ProtocolContracts memory) {
        string memory addressesJson = vm.readFile(addressesJsonPath);
        return ProtocolContracts({
            AddressManager: vm.parseJsonAddress(addressesJson, ".AddressManager"),
            L1CrossDomainMessengerProxy: vm.parseJsonAddress(addressesJson, ".L1CrossDomainMessengerProxy"),
            L1ERC721BridgeProxy: vm.parseJsonAddress(addressesJson, ".L1ERC721BridgeProxy"),
            L1StandardBridgeProxy: vm.parseJsonAddress(addressesJson, ".L1StandardBridgeProxy"),
            L2OutputOracleProxy: vm.parseJsonAddress(addressesJson, ".L2OutputOracleProxy"),
            OptimismMintableERC20FactoryProxy: vm.parseJsonAddress(addressesJson, ".OptimismMintableERC20FactoryProxy"),
            OptimismPortalProxy: vm.parseJsonAddress(addressesJson, ".OptimismPortalProxy"),
            ProxyAdmin: vm.parseJsonAddress(addressesJson, ".ProxyAdmin"),

            ProxyAdminOwner: proxyAdminOwnerExceptions[addressesJsonPath] == address(0)? controllers.FoundationMultisig : proxyAdminOwnerExceptions[addressesJsonPath],
            Challenger: challengerExceptions[addressesJsonPath],
            Guardian: guardianExceptions[addressesJsonPath] == address(0)? controllers.FoundationMultisig : guardianExceptions[addressesJsonPath]
            });
    }

    function initializeControllers() internal {
        controllers = ProtocolControllers({
            FoundationMultisig: 0x9BA6e03D8B90dE867373Db8cF1A58d2F7F006b3A,

            BaseOpsMultisig: 0x14536667Cd30e52C0b458BaACcB9faDA7046E056,
            BaseChallenger1of2: 0x6F8C5bA3F59ea3E76300E3BEcDC231D656017824,
            BaseUpgradeMultisig: 0x7bB41C3008B3f03FE483B28b8DB90e19Cf07595c,

            PgnOpsMultisig: 0x39E13D1AB040F6EA58CE19998edCe01B3C365f84,
            PgnUpgradeMultisig: 0x4a4962275DF8C60a80d3a25faEc5AA7De116A746,

            ZoraChallengerMultisig: 0xcA4571b1ecBeC86Ea2E660d242c1c29FcB55Dc72,
            ZoraGuardianMultisig: 0xC72aE5c7cc9a332699305E29F68Be66c73b60542,
            ZoraUpgradeMultisig: 0xC72aE5c7cc9a332699305E29F68Be66c73b60542
            });
    }

    function initializeExceptions() internal {
        proxyAdminOwnerExceptions["superchain/extra/addresses/mainnet/base.json"] = controllers.BaseUpgradeMultisig;
        challengerExceptions["superchain/extra/addresses/mainnet/base.json"] = controllers.BaseChallenger1of2;
        guardianExceptions["superchain/extra/addresses/mainnet/base.json"] = controllers.BaseOpsMultisig;

        challengerExceptions["superchain/extra/addresses/mainnet/op.json"] = controllers.FoundationMultisig;

        proxyAdminOwnerExceptions["superchain/extra/addresses/mainnet/pgn.json"] = controllers.PgnUpgradeMultisig;
        challengerExceptions["superchain/extra/addresses/mainnet/pgn.json"] = controllers.PgnOpsMultisig;
        guardianExceptions["superchain/extra/addresses/mainnet/pgn.json"] = controllers.PgnOpsMultisig;

        proxyAdminOwnerExceptions["superchain/extra/addresses/mainnet/zora.json"] = controllers.ZoraUpgradeMultisig;
        challengerExceptions["superchain/extra/addresses/mainnet/zora.json"] = controllers.ZoraChallengerMultisig;
        guardianExceptions["superchain/extra/addresses/mainnet/zora.json"] = controllers.ZoraGuardianMultisig;
    }
}
