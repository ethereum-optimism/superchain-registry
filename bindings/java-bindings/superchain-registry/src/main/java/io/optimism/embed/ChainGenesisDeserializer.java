package io.optimism.embed;


import com.fasterxml.jackson.core.JacksonException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;
import com.fasterxml.jackson.databind.JsonNode;
import io.optimism.superchain.BlockID;
import io.optimism.superchain.ChainGenesis;
import io.optimism.superchain.SystemConfig;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;

import java.io.IOException;
import java.util.Optional;

/**
 * The type ChainGenesis deserializer.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class ChainGenesisDeserializer extends JsonDeserializer<ChainGenesis> {
    @Override
    public ChainGenesis deserialize(JsonParser p, DeserializationContext ctxt) throws IOException, JacksonException {
        JsonNode node = p.getCodec().readTree(p);
        ChainGenesis genesis = new ChainGenesis();
        BlockID l1 = new BlockID();
        l1.setHash(Hash.fromHexString(node.get("l1").get("hash").asText()));
        l1.setNumber(node.get("l1").get("number").bigIntegerValue());
        genesis.setL1(l1);
        BlockID l2 = new BlockID();
        l2.setHash(Hash.fromHexString(node.get("l2").get("hash").asText()));
        l2.setNumber(node.get("l2").get("number").bigIntegerValue());
        genesis.setL2(l2);
        genesis.setL2Time(node.get("l2_time").asLong());
        if (node.get("extra_data") != null) {
            genesis.setExtraData(Optional.of(node.get("extra_data").asText()));
        }

        return genesis;
    }
}
