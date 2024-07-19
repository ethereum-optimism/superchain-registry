//! Contains the full superchain data.

use super::{Chain, ChainConfig, ChainList, HashMap, RollupConfigs, SuperchainConfig};

/// A list of chain configs.
#[derive(Debug, Clone, Default, Eq, PartialEq, serde::Serialize, serde::Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ChainConfigList {
    /// Chain configs.
    pub configs: Vec<ChainConfig>,
}

/// A Chain Definition.
#[derive(Debug, Clone, Default, Eq, PartialEq, serde::Serialize, serde::Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Superchain {
    /// The list of chains.
    pub chains: Vec<Chain>,
    /// Map of chain IDs to their chain configuration.
    pub op_chains: HashMap<u64, ChainConfig>,
    /// Map of chain IDs to their rollup configurations.
    pub rollup_configs: RollupConfigs,
}

impl Superchain {
    /// Initialize the superchain configurations from the chain list.
    pub fn from_chain_list() -> Self {
        let chain_list = include_str!("../../../chainList.toml");
        let chains: ChainList = toml::from_str(chain_list).unwrap();

        let mut op_chains = HashMap::new();
        let mut rollup_configs = HashMap::new();

        let mainnet_sc = include_str!("../../../superchain/configs/mainnet/superchain.toml");
        let mainnet_superchain_entry: SuperchainConfig =
            toml::from_str(mainnet_sc).expect("Failed to read mainnet superchain toml");
        let sepolia_sc = include_str!("../../../superchain/configs/sepolia/superchain.toml");
        let sepolia_superchain_entry: SuperchainConfig =
            toml::from_str(sepolia_sc).expect("Failed to read sepolia superchain toml");
        let devnet_sc = include_str!("../../../superchain/configs/sepolia-dev-0/superchain.toml");
        let devnet_superchain_entry: SuperchainConfig =
            toml::from_str(devnet_sc).expect("Failed to read devnet superchain toml");

        let config_list = include_str!("../../../superchain/configs/configs.toml");
        let configs: ChainConfigList = toml::from_str(config_list).unwrap();

        for mut config in configs.configs.into_iter() {
            // Get the chain from the chains list for the chain ID.
            let chain = chains
                .chains
                .iter()
                .find(|c| c.chain_id == config.chain_id)
                .expect("Chain not found in chain list");
            config.l1_chain_id = chain.parent.chain_id();
            if let Some(a) = &mut config.addresses {
                a.zero_proof_addresses();
            }

            let mut rollup = superchain_primitives::load_op_stack_rollup_config(&config);
            rollup.protocol_versions_address = match rollup.l1_chain_id {
                1 => mainnet_superchain_entry
                    .protocol_versions_addr
                    .unwrap_or(rollup.protocol_versions_address),
                11155111 => sepolia_superchain_entry
                    .protocol_versions_addr
                    .unwrap_or(rollup.protocol_versions_address),
                _ => devnet_superchain_entry
                    .protocol_versions_addr
                    .unwrap_or(rollup.protocol_versions_address),
            };

            rollup_configs.insert(config.chain_id, rollup);
            op_chains.insert(config.chain_id, config);
        }

        Self {
            chains: chains.chains,
            op_chains,
            rollup_configs,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use alloy_primitives::{address, b256, uint};
    use superchain_primitives::{
        AddressList, BlockID, ChainGenesis, HardForkConfiguration, SuperchainLevel, SystemConfig,
    };

    #[test]
    fn test_read_chain_configs() {
        let superchains = Superchain::from_chain_list();
        assert!(superchains.chains.len() > 1);
        let base_config = ChainConfig {
            name: String::from("Base"),
            chain_id: 8453,
            l1_chain_id: 1,
            superchain_time: Some(0),
            public_rpc: String::from("https://mainnet.base.org"),
            sequencer_rpc: String::from("https://mainnet-sequencer.base.org"),
            explorer: String::from("https://explorer.base.org"),
            superchain_level: SuperchainLevel::Frontier,
            batch_inbox_addr: address!("ff00000000000000000000000000000000008453"),
            genesis: ChainGenesis {
                l1: BlockID {
                    number: 17481768,
                    hash: b256!("5c13d307623a926cd31415036c8b7fa14572f9dac64528e857a470511fc30771"),
                },
                l2: BlockID {
                    number: 0,
                    hash: b256!("f712aa9241cc24369b143cf6dce85f0902a9731e70d66818a3a5845b296c73dd"),
                },
                l2_time: 1686789347,
                extra_data: None,
                system_config: Some(SystemConfig {
                    batcher_address: address!("5050F69a9786F081509234F1a7F4684b5E5b76C9"),
                    overhead: uint!(0xbc_U256),
                    scalar: uint!(0xa6fe0_U256),
                    gas_limit: 30000000_u64,
                    ..Default::default()
                }),
            },
            superchain: String::from(""),
            chain: String::from(""),
            hardfork_configuration: HardForkConfiguration {
                canyon_time: Some(1704992401),
                delta_time: Some(1708560000),
                ecotone_time: Some(1710374401),
                fjord_time: Some(1720627201),
                interop_time: None,
            },
            plasma: None,
            addresses: Some(AddressList {
                address_manager: address!("8EfB6B5c4767B09Dc9AA6Af4eAA89F749522BaE2"),
                l1_cross_domain_messenger_proxy: address!(
                    "866E82a600A1414e583f7F13623F1aC5d58b0Afa"
                ),
                l1_erc721_bridge_proxy: address!("608d94945A64503E642E6370Ec598e519a2C1E53"),
                l1_standard_bridge_proxy: address!("3154Cf16ccdb4C6d922629664174b904d80F2C35"),
                l2_output_oracle_proxy: Some(address!("56315b90c40730925ec5485cf004d835058518A0")),
                optimism_mintable_erc20_factory_proxy: address!(
                    "05cc379EBD9B30BbA19C6fA282AB29218EC61D84"
                ),
                optimism_portal_proxy: address!("49048044D57e1C92A77f79988d21Fa8fAF74E97e"),
                system_config_proxy: address!("73a79Fab69143498Ed3712e519A88a918e1f4072"),
                system_config_owner: address!("14536667Cd30e52C0b458BaACcB9faDA7046E056"),
                proxy_admin: address!("0475cBCAebd9CE8AfA5025828d5b98DFb67E059E"),
                proxy_admin_owner: address!("7bB41C3008B3f03FE483B28b8DB90e19Cf07595c"),
                challenger: Some(address!("6F8C5bA3F59ea3E76300E3BEcDC231D656017824")),
                guardian: address!("09f7150d8c019bef34450d6920f6b3608cefdaf2"),
                ..Default::default()
            }),
        };
        assert_eq!(*superchains.op_chains.get(&8453).unwrap(), base_config);
    }

    #[test]
    fn test_read_rollup_configs() {
        use superchain_primitives::OP_MAINNET_CONFIG;
        let superchains = Superchain::from_chain_list();
        assert_eq!(
            *superchains.rollup_configs.get(&10).unwrap(),
            OP_MAINNET_CONFIG
        );
    }
}
