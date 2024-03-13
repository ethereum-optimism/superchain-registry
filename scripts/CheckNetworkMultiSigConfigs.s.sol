// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import {console2} from "forge-std/console2.sol";
import {Script} from "forge-std/Script.sol";
import {VmSafe} from "forge-std/Vm.sol";

/**
 * @title CheckNetworkMultiSigConfigs
 * @notice A script to check security configurations of important multisigs for a given network.
 *         The usage is as follows:
 *         $ forge script CheckNetworkMultiSigConfigs --rpc-url $MAINNET_RPC_URL
 */
contract CheckNetworkMultiSigConfigs is Script {
    struct MultiSigConfig {
        address multiSigAddress;
        uint256 threshold;
        uint256 totalSigners;
        address[] signers;
    }

    MultiSigConfig[] multiSigConfigs;

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
                    "Unsupported chain ID: ",
                    vm.toString(block.chainid),
                    ". Please call runOnDir(string,bool) directly."
                )
            );
        }
        string memory jsonDir = string.concat("superchain/extra/addresses/", network);
        runOnDir(jsonDir);
    }

    function runOnDir(string memory jsonDir) public {
        string memory multiSigsFile = vm.readFile(string.concat(jsonDir, "/multiSigs.json"));
        hasErrors = false;

        MultiSigConfig memory securityCouncilConfig = getMultiSigConfig(multiSigsFile, ".SecurityCouncil");
        console2.log("\nChecking Security Council MultiSig: %s", securityCouncilConfig.multiSigAddress);
        checkMultiSigConfig(securityCouncilConfig);
        console2.log("Security Council MultiSig Check Complete");

        MultiSigConfig memory opFoundationConfig = getMultiSigConfig(multiSigsFile, ".OpFoundation");
        console2.log("\nChecking Op Foundation MultiSig: %s", opFoundationConfig.multiSigAddress);
        checkMultiSigConfig(opFoundationConfig);
        console2.log("Op Foundation MultiSig Check Complete");

        require(!hasErrors, "Errors occurred: See logs above for more info");
    }

    function getMultiSigConfig(string memory jsonData, string memory jsonProperty)
        internal
        view
        returns (MultiSigConfig memory)
    {
        // TODO: Fix: This reverted when trying to marshell the signers address array? Decoding manually for now.
        // MultiSigConfig memory multiSigConfig = abi.decode(securityCouncilDetails, (MultiSigConfig));

        address multiSigAddress = vm.parseJsonAddress(jsonData, string.concat(jsonProperty, ".multiSigAddress"));
        uint256 threshold = vm.parseJsonUint(jsonData, string.concat(jsonProperty, ".threshold"));
        uint256 totalSigners = vm.parseJsonUint(jsonData, string.concat(jsonProperty, ".totalSigners"));
        address[] memory signers = vm.parseJsonAddressArray(jsonData, string.concat(jsonProperty, ".signers"));

        if (signers.length != totalSigners) {
            console2.log(unicode"❌ Error: %s != %s", signers.length, totalSigners);
            revert("Invalid input data for number of signers.");
        }

        return MultiSigConfig({
            multiSigAddress: multiSigAddress,
            threshold: threshold,
            totalSigners: totalSigners,
            signers: signers
        });
    }

    function checkMultiSigConfig(MultiSigConfig memory multiSigConfig) internal {
        string memory getOwnersSignature = "getOwners()";
        string memory thresholdSignature = "getThreshold()";
        address[] memory networkMultiSigOwners = getMultiSigOwners(multiSigConfig, getOwnersSignature);
        compareMultiSigOwners(networkMultiSigOwners, multiSigConfig);
        checkThreshold(multiSigConfig, thresholdSignature);
    }

    function checkThreshold(MultiSigConfig memory multiSigConfig, string memory signature)
        internal
        returns (uint256 threshold)
    {
        vm.prank(address(0));
        (bool success, bytes memory thresholdBytes) =
            multiSigConfig.multiSigAddress.staticcall(abi.encodeWithSignature(signature));

        if (!success) {
            revert("Error getting threshold for multiSig.");
        }

        uint256 networkThreshold = abi.decode(thresholdBytes, (uint256));

        if (networkThreshold != multiSigConfig.threshold) {
            hasErrors = true;
            console2.log(unicode"❌ Error: %s != %s.%s", multiSigConfig.threshold, multiSigConfig.multiSigAddress, signature);
            console2.log("Actual threshold: %s", networkThreshold);
        }
        return networkThreshold;
    }

    function getMultiSigOwners(MultiSigConfig memory multiSigConfig, string memory signature)
        internal
        returns (address[] memory owners)
    {
        vm.prank(address(0));
        (bool success, bytes memory addrBytes) =
            multiSigConfig.multiSigAddress.staticcall(abi.encodeWithSignature(signature));

        if (!success) {
            revert("Error getting multiSig owners.");
        }

        address[] memory networkMultiSigOwners = abi.decode(addrBytes, (address[]));
        if (networkMultiSigOwners.length != multiSigConfig.totalSigners) {
            hasErrors = true;
            console2.log(
                unicode"❌ Error: %s != %s.%s", multiSigConfig.totalSigners, multiSigConfig.multiSigAddress, signature
            );
            console2.log("Actual number of signers: %s", networkMultiSigOwners.length);
        }
        return networkMultiSigOwners;
    }

    /**
     * Sorting both sets of owners before comparing
     */
    function compareMultiSigOwners(address[] memory networkOwners, MultiSigConfig memory multiSigConfig)
        internal
        returns (bool)
    {
        address[] memory sortedNetworkOwners = bubbleSort(networkOwners);
        address[] memory sortedLocalOwners = bubbleSort(multiSigConfig.signers);

        for (uint256 i = 0; i < sortedNetworkOwners.length; i++) {
            if (sortedNetworkOwners[i] != sortedLocalOwners[i]) {
                hasErrors = true;
                console2.log(unicode"❌ Error: %s != %s", sortedNetworkOwners[i], sortedLocalOwners[i]);
                return false;
            }
        }
        return true;
    }

    function bubbleSort(address[] memory arr) public pure returns (address[] memory) {
        uint256 n = arr.length;
        for (uint256 i = 0; i < n - 1; i++) {
            for (uint256 j = 0; j < n - i - 1; j++) {
                if (arr[j] > arr[j + 1]) {
                    (arr[j], arr[j + 1]) = (arr[j + 1], arr[j]);
                }
            }
        }
        return arr;
    }
}
