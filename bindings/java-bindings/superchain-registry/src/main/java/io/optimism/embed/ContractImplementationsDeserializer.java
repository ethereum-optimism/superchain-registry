package io.optimism.embed;

import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.JsonNode;
import com.google.common.base.Strings;
import io.optimism.superchain.ContractImplementations;
import org.hyperledger.besu.datatypes.Address;

import java.io.IOException;
import java.math.BigInteger;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

/**
 * The type ContractImplementations deserializer.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class ContractImplementationsDeserializer extends JsonDeserializer<ContractImplementations> {
    @Override
    public ContractImplementations deserialize(JsonParser p, DeserializationContext ctxt) throws IOException {
        JsonNode node = p.getCodec().readTree(p);
        ContractImplementations contractImplementations = new ContractImplementations();

        node.fields().forEachRemaining(entry -> {
            switch (entry.getKey()) {
                case "l1_cross_domain_messenger":
                    entry.getValue().fields().forEachRemaining(addressEntry -> contractImplementations.getL1CrossDomainMessenger().put(addressEntry.getKey(), getAddress(addressEntry)));
                    break;
                case "l1_erc721_bridge":
                    entry.getValue().fields().forEachRemaining(addressEntry -> contractImplementations.getL1ERC721Bridge().put(addressEntry.getKey(), getAddress(addressEntry)));
                    break;
                case "l1_standard_bridge":
                    entry.getValue().fields().forEachRemaining(addressEntry -> contractImplementations.getL1StandardBridge().put(addressEntry.getKey(), getAddress(addressEntry)));
                    break;
                case "l2_output_oracle":
                    entry.getValue().fields().forEachRemaining(addressEntry -> contractImplementations.getL2OutputOracle().put(addressEntry.getKey(), getAddress(addressEntry)));
                    break;
                case "optimism_mintable_erc20_factory":
                    entry.getValue().fields().forEachRemaining(addressEntry -> contractImplementations.getOptimismMintableERC20Factory().put(addressEntry.getKey(), getAddress(addressEntry)));
                    break;
                case "optimism_portal":
                    entry.getValue().fields().forEachRemaining(addressEntry -> contractImplementations.getOptimismPortal().put(addressEntry.getKey(), getAddress(addressEntry)));
                    break;
                case "system_config":
                    entry.getValue().fields().forEachRemaining(addressEntry -> contractImplementations.getSystemConfig().put(addressEntry.getKey(), getAddress(addressEntry)));
                    break;
                case "anchor_state_registry":
                    entry.getValue().fields().forEachRemaining(addressEntry -> {
                        contractImplementations.setAnchorStateRegistry(Optional.of(new HashMap<>()));
                        contractImplementations.getAnchorStateRegistry().orElseThrow().put(addressEntry.getKey(), getAddress(addressEntry));
                    });
                    break;
                case "delayed_weth":
                    entry.getValue().fields().forEachRemaining(addressEntry -> {
                        contractImplementations.setDelayedWETH(Optional.of(new HashMap<>()));
                        contractImplementations.getDelayedWETH().orElseThrow().put(addressEntry.getKey(), getAddress(addressEntry));
                    });
                    break;
                case "dispute_game_factory":
                    entry.getValue().fields().forEachRemaining(addressEntry -> {
                        contractImplementations.setDisputeGameFactory(Optional.of(new HashMap<>()));
                        contractImplementations.getDisputeGameFactory().orElseThrow().put(addressEntry.getKey(), getAddress(addressEntry));
                    });
                    break;
                case "fault_dispute_game":
                    entry.getValue().fields().forEachRemaining(addressEntry -> {
                        contractImplementations.setFaultDisputeGame(Optional.of(new HashMap<>()));
                        contractImplementations.getFaultDisputeGame().orElseThrow().put(addressEntry.getKey(), getAddress(addressEntry));
                    });
                    break;
                case "mips":
                    entry.getValue().fields().forEachRemaining(addressEntry -> {
                        contractImplementations.setMips(Optional.of(new HashMap<>()));
                        contractImplementations.getMips().orElseThrow().put(addressEntry.getKey(), getAddress(addressEntry));
                    });
                    break;
                case "permissioned_dispute_game":
                    entry.getValue().fields().forEachRemaining(addressEntry -> {
                        contractImplementations.setPermissionedDisputeGame(Optional.of(new HashMap<>()));
                        contractImplementations.getPermissionedDisputeGame().orElseThrow().put(addressEntry.getKey(), getAddress(addressEntry));
                    });
                    break;
                case "preimage_oracle":
                    entry.getValue().fields().forEachRemaining(addressEntry -> {
                        contractImplementations.setPreimageOracle(Optional.of(new HashMap<>()));
                        contractImplementations.getPreimageOracle().orElseThrow().put(addressEntry.getKey(), getAddress(addressEntry));
                    });
                    break;
                default:
                    break;
            }
        });

        return contractImplementations;
    }

    private static Address getAddress(Map.Entry<String, JsonNode> entry) {
        return Address.fromHexStringStrict(Strings.padStart(new BigInteger(entry.getValue().asText()).toString(16), 40, '0'));
    }

}
