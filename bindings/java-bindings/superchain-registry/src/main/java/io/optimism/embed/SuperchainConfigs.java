package io.optimism.embed;


import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.fasterxml.jackson.databind.module.SimpleModule;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.fasterxml.jackson.datatype.jdk8.Jdk8Module;
import com.google.common.base.Strings;
import com.google.common.io.Resources;
import io.optimism.superchain.AddressList;
import io.optimism.superchain.ChainConfig;
import io.optimism.superchain.ChainGenesis;
import io.optimism.superchain.ContractImplementations;
import io.optimism.superchain.Superchain;
import io.optimism.superchain.SuperchainConfig;
import io.optimism.superchain.SuperchainLevel;
import io.optimism.superchain.SystemConfig;

import java.io.IOException;
import java.net.URISyntaxException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Consumer;
import java.util.stream.Stream;

public final class SuperchainConfigs {

    private static final Path CONFIGS_PATH;

    private static final Path EXTRA_PATH;

    private static final Path IMPLEMENTATIONS_PATH;

    private static final ObjectMapper UPPER_CAMEL_CASE_MAPPER = new ObjectMapper();

    private static final ObjectMapper LOWER_CAMEL_CASE_MAPPER = new ObjectMapper();

    private static final ObjectMapper SNAKE_CASE_MAPPER = new ObjectMapper();

    private static final ObjectMapper YAML_MAPPER = new ObjectMapper(new YAMLFactory());

    private static final ObjectMapper JSON_MAPPER = new ObjectMapper();

    static {
        try {
            CONFIGS_PATH = Paths.get(Resources.getResource("configs").toURI());
            EXTRA_PATH = Paths.get(Resources.getResource("extra").toURI());
            IMPLEMENTATIONS_PATH = Paths.get(Resources.getResource("implementations").toURI());

            UPPER_CAMEL_CASE_MAPPER.setPropertyNamingStrategy(PropertyNamingStrategies.UpperCamelCaseStrategy.INSTANCE);
            LOWER_CAMEL_CASE_MAPPER.setPropertyNamingStrategy(PropertyNamingStrategies.LowerCamelCaseStrategy.INSTANCE);
            SNAKE_CASE_MAPPER.setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);

            JSON_MAPPER.registerModule(new Jdk8Module());
            YAML_MAPPER.registerModule(new Jdk8Module());
            SimpleModule yamlModule = new SimpleModule();
            yamlModule.addDeserializer(SuperchainConfig.class, new SuperchainConfigDeserializer());
            yamlModule.addDeserializer(ContractImplementations.class, new ContractImplementationsDeserializer());
            yamlModule.addDeserializer(ChainGenesis.class, new ChainGenesisDeserializer());
            yamlModule.addDeserializer(SuperchainLevel.class, new SuperchainLevelDeserializer());
            yamlModule.addDeserializer(ChainConfig.class, new ChainConfigDeserializer());
            YAML_MAPPER.registerModule(yamlModule);

            SimpleModule jsonModule = new SimpleModule();
            jsonModule.addDeserializer(AddressList.class, new AddressListDeserializer());
            JSON_MAPPER.registerModule(jsonModule);
        } catch (URISyntaxException e) {
            throw new RuntimeException(e);
        }
    }

    public static final Map<String, Superchain> SUPERCHAINS = new HashMap<>();

    public static final Map<Long, ChainConfig> OP_CHAINS = new HashMap<>();

    public static final Map<Long, AddressList> ADDRESSES = new HashMap<>();

    public static final Map<Long, SystemConfig> GENESIS_SYSTEM_CONFIGS = new HashMap<>();

    public static final Map<String, ContractImplementations> IMPLEMENTATIONS = new HashMap<>();

    public static void loadConfigs() throws IOException {
        final SuperchainConfig[] superchainConfig = {new SuperchainConfig()};
        try (Stream<Path> pathStream = Files.list(CONFIGS_PATH)) {
            pathStream.filter(Files::isDirectory).forEach(new Consumer<Path>() {
                @Override
                public void accept(Path path) {
                    String network = path.getFileName().toString();
                    List<Long> chainIDs = new ArrayList<>();
                    try (Stream<Path> fileStream = Files.list(path)) {
                        fileStream.filter(Files::isRegularFile).forEach(new Consumer<Path>() {
                            @Override
                            public void accept(Path path) {
                                try {
                                    String content = Files.readString(path, StandardCharsets.UTF_8);
                                    String fileName = path.getFileName().toString();
                                    if (fileName.equals("superchain.yaml")) {
                                        YAML_MAPPER.setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
                                        superchainConfig[0] = YAML_MAPPER.readValue(content, SuperchainConfig.class);
                                        System.out.println(superchainConfig[0]);
                                        return;
                                    }
                                    if (fileName.equals("semver.yaml")) {
                                        // TODO
                                        return;
                                    }

                                    String chainName = fileName.replaceFirst("[.][^.]+$", "");
                                    YAML_MAPPER.setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
                                    ChainConfig chainConfig = YAML_MAPPER.readValue(content, ChainConfig.class);
                                    chainConfig.setChain(chainName);
                                    chainConfig.setMissingForkConfigs(superchainConfig[0].getHardforkDefaults());
                                    System.out.println(chainConfig);

                                    String jsonFileName = chainConfig.getChain() + ".json";
                                    Path addressesDataFile = Paths.get(String.valueOf(EXTRA_PATH), "addresses", network, jsonFileName);
                                    String addressesData = Files.readString(addressesDataFile, StandardCharsets.UTF_8);
                                    JSON_MAPPER.setPropertyNamingStrategy(PropertyNamingStrategies.UPPER_CAMEL_CASE);
                                    AddressList addressList = JSON_MAPPER.readValue(addressesData, AddressList.class);
                                    System.out.println(addressList);

                                    Path genesisDataFile = Paths.get(String.valueOf(EXTRA_PATH), "genesis-system-configs", network, jsonFileName);
                                    String genesisData = Files.readString(genesisDataFile, StandardCharsets.UTF_8);
                                    JSON_MAPPER.setPropertyNamingStrategy(PropertyNamingStrategies.LOWER_CAMEL_CASE);
                                    SystemConfig systemConfig = JSON_MAPPER.readValue(genesisData, SystemConfig.class);
                                    System.out.println(systemConfig);

                                    long id = chainConfig.getChainId();
                                    chainConfig.setSuperchain(network);
                                    GENESIS_SYSTEM_CONFIGS.put(id, systemConfig);
                                    ADDRESSES.put(id, addressList);
                                    OP_CHAINS.put(id, chainConfig);
                                    chainIDs.add(id);
                                } catch (IOException e) {
                                    e.printStackTrace();
                                }
                            }
                        });
                    } catch (IOException e) {
                        e.printStackTrace();
                    }

                    switch (network) {
                        case "mainnet":
                            if (!Strings.isNullOrEmpty(System.getenv("CIRCLE_CI_MAINNET_RPC"))) {
                                superchainConfig[0].getL1().setPublicRpc(System.getenv("CIRCLE_CI_MAINNET_RPC"));
                            }
                            break;
                        case "sepolia-dev-0", "sepolia":
                            if (!Strings.isNullOrEmpty(System.getenv("CIRCLE_CI_SEPOLIA_RPC"))) {
                                superchainConfig[0].getL1().setPublicRpc(System.getenv("CIRCLE_CI_SEPOLIA_RPC"));
                            }
                            break;
                        default:
                            break;
                    }

                    try {
                        ContractImplementations contractImplementations = newContractImplementations(network);
                        IMPLEMENTATIONS.put(network, contractImplementations);
                        System.out.println(IMPLEMENTATIONS);

                        SUPERCHAINS.put(network, new Superchain(superchainConfig[0], chainIDs, network));
                        System.out.println(SUPERCHAINS);
                    } catch (IOException e) {
                        throw new RuntimeException(e);
                    }
                }
            });
        }


    }

    private static ContractImplementations newContractImplementations(String network) throws IOException {
        String implementationsData = Files.readString(Paths.get(String.valueOf(IMPLEMENTATIONS_PATH), "implementations.yaml"), StandardCharsets.UTF_8);
        YAML_MAPPER.setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
        ContractImplementations globalContractImplementations = YAML_MAPPER.readValue(implementationsData, ContractImplementations.class);

        if (Strings.isNullOrEmpty(network)) {
            return globalContractImplementations;
        }

        String networkImplementationsData = Files.readString(Paths.get(String.valueOf(IMPLEMENTATIONS_PATH), "networks", String.format("%s.yaml", network)), StandardCharsets.UTF_8);
        ContractImplementations networkContractImplementations = YAML_MAPPER.readValue(networkImplementationsData, ContractImplementations.class);

        globalContractImplementations.merge(networkContractImplementations);

        return globalContractImplementations;
    }
}
