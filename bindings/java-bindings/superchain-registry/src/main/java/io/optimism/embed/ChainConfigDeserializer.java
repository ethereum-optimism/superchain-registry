package io.optimism.embed;

import com.fasterxml.jackson.core.JacksonException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.JsonNode;
import io.optimism.superchain.ChainConfig;
import io.optimism.superchain.ChainGenesis;
import io.optimism.superchain.HardForkConfiguration;
import io.optimism.superchain.PlasmaConfig;
import io.optimism.superchain.SuperchainLevel;
import org.hyperledger.besu.datatypes.Address;

import java.io.IOException;
import java.util.Optional;

/**
 * Created by IntelliJ IDEA.
 * Author: kaichen
 * Date: 2024/6/23
 * Time: 21:54
 */
public class ChainConfigDeserializer extends JsonDeserializer<ChainConfig> {
    @Override
    public ChainConfig deserialize(JsonParser p, DeserializationContext ctxt) throws IOException, JacksonException {
        JsonNode node = p.getCodec().readTree(p);
        ChainConfig config = new ChainConfig();
        config.setName(node.get("name").asText());
        config.setChainId(node.get("chain_id").asLong());
        config.setPublicRpc(node.get("public_rpc").asText());
        config.setSequencerRpc(node.get("sequencer_rpc").asText());
        config.setExplorer(node.get("explorer").asText());
        config.setSuperchainLevel(SuperchainLevel.fromValue((byte) node.get("superchain_level").asInt()));
        if (node.get("superchain_time") != null) {
            config.setSuperchainTime(Optional.of(node.get("superchain_time").asLong()));
        }
        config.setBatchInboxAddr(Address.fromHexStringStrict(node.get("batch_inbox_addr").asText()));
        config.setGenesis(p.getCodec().treeToValue(node.get("genesis"), ChainGenesis.class));

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

        config.setHardforkConfiguration(hardForkConfiguration);

        if (node.get("plasma") != null) {
            config.setPlasma(Optional.of(p.getCodec().treeToValue(node.get("plasma"), PlasmaConfig.class)));
        }
        return config;
    }
}
