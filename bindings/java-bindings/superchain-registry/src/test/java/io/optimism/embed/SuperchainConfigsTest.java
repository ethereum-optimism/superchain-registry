package io.optimism.embed;

import org.junit.jupiter.api.Test;

import java.io.IOException;


class SuperchainConfigsTest {

    @Test
    void loadConfigs() throws IOException {
        SuperchainConfigs.loadConfigs();
    }
}
