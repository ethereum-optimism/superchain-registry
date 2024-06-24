package io.optimism.embed;

import com.fasterxml.jackson.core.JacksonException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.JsonNode;
import com.google.common.base.Strings;
import io.optimism.superchain.HardForkConfiguration;
import io.optimism.superchain.SuperchainConfig;
import io.optimism.superchain.SuperchainL1Info;

import java.io.IOException;
import java.util.Optional;

import static org.hyperledger.besu.datatypes.Address.fromHexStringStrict;

/**
 * The type SuperchainConfig desirializer.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class SuperchainConfigDeserializer extends JsonDeserializer<SuperchainConfig> {
    @Override
    public SuperchainConfig deserialize(JsonParser p, DeserializationContext ctxt) throws IOException, JacksonException {
        JsonNode node = p.getCodec().readTree(p);
        SuperchainConfig config = new SuperchainConfig();
        config.setName(node.get("name").asText());
        SuperchainL1Info l1 = new SuperchainL1Info();
        l1.setChainId(node.get("l1").get("chain_id").asLong());
        l1.setPublicRpc(node.get("l1").get("public_rpc").asText());
        l1.setExplorer(node.get("l1").get("explorer").asText());
        config.setL1(l1);
        if (Strings.isNullOrEmpty(node.get("protocol_versions_addr").asText())) {
            config.setProtocolVersionsAddr(Optional.empty());
        } else {
            config.setProtocolVersionsAddr(Optional.of(fromHexStringStrict(node.get("protocol_versions_addr").asText())));
        }
        if (Strings.isNullOrEmpty(node.get("superchain_config_addr").asText())) {
            config.setSuperchainConfigAddr(Optional.empty());
        } else {
            config.setSuperchainConfigAddr(Optional.of(fromHexStringStrict(node.get("superchain_config_addr").asText())));
        }
        HardForkConfiguration hardForkConfiguration = new HardForkConfiguration();
        if (node.get("canyon_time") != null) {
            hardForkConfiguration.setCanyonTime(Optional.of(node.get("canyon_time").asLong()));
        } else {
            hardForkConfiguration.setCanyonTime(Optional.empty());
        }
        if (node.get("delta_time") != null) {
            hardForkConfiguration.setDeltaTime(Optional.of(node.get("delta_time").asLong()));
        } else {
            hardForkConfiguration.setDeltaTime(Optional.empty());
        }
        if (node.get("ecotone_time") != null) {
            hardForkConfiguration.setEcotoneTime(Optional.of(node.get("ecotone_time").asLong()));
        } else {
            hardForkConfiguration.setEcotoneTime(Optional.empty());
        }
        if (node.get("fjord_time") != null) {
            hardForkConfiguration.setFjordTime(Optional.of(node.get("fjord_time").asLong()));
        } else {
            hardForkConfiguration.setFjordTime(Optional.empty());
        }
        config.setHardforkDefaults(hardForkConfiguration);
        return config;
    }
}
