//! Contains the full superchain data.

use super::{
    AddressList, Addresses, ChainConfig, GenesisSystemConfigs, Implementations, OPChains,
    RollupConfigs, HashMap, Chain, SystemConfig, SuperchainConfig
};

/// A Chain Definition.
#[derive(Debug, Clone, Default, Eq, PartialEq, serde::Serialize, serde::Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Superchain {
    /// The list of chains.
    pub chains: Vec<Chain>,
    /// Map of chain IDs to their chain configuration.
    pub op_chains: OPChains,
    /// Map of chain IDs to their address list.
    pub addresses: Addresses,
    /// Map of chain IDs to their genesis system configurations.
    pub genesis_system_configs: GenesisSystemConfigs,
    /// Map of superchain names to their contract implementations.
    pub implementations: Implementations,
    /// Map of chain IDs to their rollup configurations.
    pub rollup_configs: RollupConfigs,
}

impl Superchain {
    /// Returns raw chain config bytes.
    pub(crate) fn raw_chain_configs() -> Vec<String> {
        vec![
            // Mainnet
            include_str!("../../../superchain/configs/mainnet/base.yaml").to_string(),
            include_str!("../../../superchain/configs/mainnet/lyra.yaml").to_string(),
            include_str!("../../../superchain/configs/mainnet/metal.yaml").to_string(),
            include_str!("../../../superchain/configs/mainnet/mode.yaml").to_string(),
            include_str!("../../../superchain/configs/mainnet/op.yaml").to_string(),
            include_str!("../../../superchain/configs/mainnet/orderly.yaml").to_string(),
            include_str!("../../../superchain/configs/mainnet/pgn.yaml").to_string(),
            include_str!("../../../superchain/configs/mainnet/zora.yaml").to_string(),
            // Sepolia
            include_str!("../../../superchain/configs/sepolia/base.yaml").to_string(),
            include_str!("../../../superchain/configs/sepolia/mode.yaml").to_string(),
            include_str!("../../../superchain/configs/sepolia/op.yaml").to_string(),
            include_str!("../../../superchain/configs/sepolia/pgn.yaml").to_string(),
            include_str!("../../../superchain/configs/sepolia/race.yaml").to_string(),
            include_str!("../../../superchain/configs/sepolia/zora.yaml").to_string(),
            // Spolia Dev 0
            include_str!("../../../superchain/configs/sepolia-dev-0/base-devnet-0.yaml")
                .to_string(),
            include_str!("../../../superchain/configs/sepolia-dev-0/oplabs-devnet-0.yaml")
                .to_string(),
        ]
    }

    pub(crate) fn raw_genesis_configs() -> Vec<(String, String)> {
        vec![
            // Mainnet
            ("mainnet/base".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/mainnet/base.json").to_string()),
            ("mainnet/lyra".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/mainnet/lyra.json").to_string()),
            ("mainnet/metal".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/mainnet/metal.json").to_string()),
            ("mainnet/mode".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/mainnet/mode.json").to_string()),
            ("mainnet/op".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/mainnet/op.json").to_string()),
            ("mainnet/orderly".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/mainnet/orderly.json").to_string()),
            ("mainnet/pgn".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/mainnet/pgn.json").to_string()),
            ("mainnet/superlumio".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/mainnet/superlumio.json").to_string()),
            ("mainnet/zora".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/mainnet/zora.json").to_string()),
            // Sepolia
            ("sepolia/base".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/sepolia/base.json").to_string()),
            ("sepolia/metal".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/sepolia/metal.json").to_string()),
            ("sepolia/mode".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/sepolia/mode.json").to_string()),
            ("sepolia/op".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/sepolia/op.json").to_string()),
            ("sepolia/pgn".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/sepolia/pgn.json").to_string()),
            ("sepolia/race".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/sepolia/race.json").to_string()),
            ("sepolia/zora".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/sepolia/zora.json").to_string()),
            // Spolia Dev 0
            ("sepolia-dev-0/base-devnet-0".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/sepolia-dev-0/base-devnet-0.json")
                .to_string()),
            ("sepolia-dev-0/oplabs-devnet-0".to_string(), include_str!("../../../superchain/extra/genesis-system-configs/sepolia-dev-0/oplabs-devnet-0.json")
                .to_string()),
        ]
    }

    /// Initialize the superchain configurations from the chain list.
    pub fn from_chain_list() -> Self {
        let chain_list = include_str!("../../../superchain/configs/chainList.json");
        let chains: Vec<Chain> = serde_json::from_str(chain_list).unwrap();

        let mut op_chains = HashMap::new();
        let mut addresses = HashMap::new();
        let mut genesis_system_configs = HashMap::new();
        let mut rollup_configs = HashMap::new();

        let address_list = include_str!("../../../superchain/extra/addresses/addresses.json");
        let addrs: HashMap<u64, AddressList> =
            serde_json::from_str(address_list).expect("Failed to deserialize address list file");
        for (id, list) in addrs.into_iter() {
            addresses.insert(id, list);
        }

        let mainnet_sc = include_str!("../../../superchain/configs/mainnet/superchain.yaml");
        let mainnet_superchain_entry: SuperchainConfig = serde_yaml::from_str(mainnet_sc).expect("Failed to read mainnet superchain yaml");
        let sepolia_sc = include_str!("../../../superchain/configs/sepolia/superchain.yaml");
        let sepolia_superchain_entry: SuperchainConfig = serde_yaml::from_str(sepolia_sc).expect("Failed to read sepolia superchain yaml");
        let devnet_sc = include_str!("../../../superchain/configs/sepolia-dev-0/superchain.yaml");
        let devnet_superchain_entry: SuperchainConfig = serde_yaml::from_str(devnet_sc).expect("Failed to read devnet superchain yaml");

        for (chain, genesis) in Superchain::raw_genesis_configs() {
            let genesis_config: SystemConfig = serde_json::from_str(&genesis).expect("Failed to deserialize genesis system config");
            if let Some(cfg) = chains.iter().find(|c| c.identifier == chain) {
                genesis_system_configs.insert(cfg.chain_id, genesis_config);
            }
        }

        for chain in Superchain::raw_chain_configs() {
            let mut chain_config: ChainConfig =
                serde_yaml::from_str(&chain).expect("Failed to deserialize chain config file");
            chain_config.chain = chain_config.chain.trim_end_matches(".yaml").to_string();
            chain_config.genesis.system_config = genesis_system_configs.get(&chain_config.chain_id).cloned();
            let l1_chain_id = chains.iter().find(|c| c.chain_id == chain_config.chain_id).map(|c| c.parent.chain_id()).unwrap_or(10);
            match l1_chain_id {
                1 => chain_config.set_missing_fork_configs(&mainnet_superchain_entry.hardfork_defaults),
                11155111 => chain_config.set_missing_fork_configs(&sepolia_superchain_entry.hardfork_defaults),
                _ => chain_config.set_missing_fork_configs(&devnet_superchain_entry.hardfork_defaults),
            }
            chain_config.l1_chain_id = l1_chain_id;
            op_chains.insert(chain_config.chain_id, chain_config.clone());
        }

        for chain in chains.iter() {
            let chain_config = op_chains.get(&chain.chain_id).unwrap();
            let addrs = addresses.get(&chain.chain_id).unwrap();
            let rollup_config =
                superchain_primitives::load_op_stack_rollup_config(chain_config, addrs);
            rollup_configs.insert(chain.chain_id, rollup_config);
        }

        Self {
            chains,
            op_chains,
            addresses,
            genesis_system_configs,
            implementations: HashMap::new(),
            rollup_configs,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use alloy_primitives::{uint, address};

    #[test]
    fn test_read_chain_configs() {
        let superchains = Superchain::from_chain_list();
        assert!(superchains.chains.len() > 1);
        assert_eq!(superchains.op_chains.get(&8453).unwrap().name, "Base");
    }

    #[test]
    fn test_read_chain_addresses() {
        let superchains = Superchain::from_chain_list();
        let base_address_list = AddressList {
            address_manager: address!("8EfB6B5c4767B09Dc9AA6Af4eAA89F749522BaE2"),
            l1_cross_domain_messenger_proxy: address!("866E82a600A1414e583f7F13623F1aC5d58b0Afa"),
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
            guardian: address!("14536667Cd30e52C0b458BaACcB9faDA7046E056"),
            ..Default::default()
        };
        assert_eq!(
            *superchains.addresses.get(&8453).unwrap(),
            base_address_list
        );
    }

    #[test]
    fn test_read_genesis_system_configs() {
        let superchains = Superchain::from_chain_list();
        let base_sys_config = SystemConfig {
            batcher_addr: address!("5050F69a9786F081509234F1a7F4684b5E5b76C9"),
            overhead: uint!(0xbc_U256),
            scalar: uint!(0xa6fe0_U256),
            gas_limit: 30000000_u64,
            ..Default::default()
        };
        assert_eq!(
            *superchains.genesis_system_configs.get(&8453).unwrap(),
            base_sys_config
        );
    }

    #[test]
    fn test_read_rollup_configs() {
        use superchain_primitives::OP_MAINNET_CONFIG;
        let superchains = Superchain::from_chain_list();
        assert_eq!(*superchains.rollup_configs.get(&10).unwrap(), OP_MAINNET_CONFIG);
    }
}
