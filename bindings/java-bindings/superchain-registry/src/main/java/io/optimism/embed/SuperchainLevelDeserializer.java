package io.optimism.embed;

import com.fasterxml.jackson.core.JacksonException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonDeserializer;
import io.optimism.superchain.SuperchainLevel;

import java.io.IOException;

/**
 * The type Superchainlevel deserializer.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class SuperchainLevelDeserializer extends JsonDeserializer<SuperchainLevel> {
    @Override
    public SuperchainLevel deserialize(JsonParser jsonParser, DeserializationContext deserializationContext) throws IOException, JacksonException {
        return SuperchainLevel.fromValue((byte) jsonParser.getValueAsInt());
    }
}
