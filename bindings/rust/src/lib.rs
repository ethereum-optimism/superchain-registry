#![doc = include_str!("../README.md")]
// #![warn(missing_debug_implementations, missing_docs, rustdoc::all)]
#![deny(unused_must_use, rust_2018_idioms)]
#![cfg_attr(docsrs, feature(doc_cfg, doc_auto_cfg))]

use lazy_static::lazy_static;
use std::collections::HashMap;

/// Module holding the superchain configuration definition types.
pub mod superchain;
use superchain::{
    is_config_file, AddressList, Addresses, ChainConfig, ContractImplementations,
    GenesisSystemConfig, GenesisSystemConfigs, Implementations, OPChains, Superchain,
    SuperchainConfig, Superchains,
};

/// Module responsible for representing the embedded filesystem
/// that contains the superchain configurations for static access.
mod embed;

/// Module for utility functions and types.
mod util;

lazy_static! {
    /// Private initializer that runs once to load the superchain configurations.
    static ref _INIT: InitTuple = load_embedded_configs();

    /// Superchain configurations exported from the registry
    pub static ref SUPERCHAINS: Superchains = _INIT.0.clone();

    /// OPChain configurations exported from the registry
    pub static ref OPCHAINS: OPChains = _INIT.1.clone();

    /// Address lists exported from the registry
    pub static ref ADDRESSES: Addresses = _INIT.2.clone();

    /// Genesis system configurations exported from the registry
    pub static ref GENESIS_SYSTEM_CONFIGS: GenesisSystemConfigs = _INIT.3.clone();

    /// Contract implementations exported from the registry
    pub static ref IMPLEMENTATIONS: Implementations = _INIT.4.clone();
}

type InitTuple = (
    Superchains,
    OPChains,
    Addresses,
    GenesisSystemConfigs,
    Implementations,
);

/// Initialize the superchain configurations from the embedded filesystem.
fn load_embedded_configs() -> InitTuple {
    let mut superchains = HashMap::new();
    let mut op_chains = HashMap::new();
    let mut addresses = HashMap::new();
    let mut genesis_system_configs = HashMap::new();
    let mut implementations = HashMap::new();

    for target_dir in embed::CONFIGS_DIR.dirs() {
        let target_name = target_dir.path().file_name().unwrap().to_str().unwrap();
        let target_data = target_dir
            .get_file(format!("{}/superchain.yaml", target_name))
            .expect("Failed to find superchain.yaml config file")
            .contents();
        let mut entry_config: SuperchainConfig = serde_yaml::from_slice(target_data)
            .expect("Failed to deserialize superchain.yaml config file");

        let mut chain_ids = Vec::new();
        for chain in target_dir.entries() {
            let chain_name = chain.path().file_name().unwrap().to_str().unwrap();
            if !is_config_file(chain_name) {
                continue;
            }

            let chain_data = chain.as_file().unwrap().contents();
            let mut chain_config: ChainConfig = serde_yaml::from_slice(chain_data)
                .expect("Failed to deserialize chain config file");
            chain_config.chain = chain_name.trim_end_matches(".yaml").to_string();

            chain_config.set_missing_fork_configs(&entry_config.hardfork_defaults);

            let json_file_name = chain_config.chain.clone() + ".json";
            let addresses_data = embed::EXTRA_DIR
                .get_file(format!("addresses/{}/{}", target_name, json_file_name))
                .expect("Failed to find address list file")
                .contents();
            let addrs: AddressList = serde_json::from_slice(addresses_data)
                .expect("Failed to deserialize address list file");

            let genesis_config_data = embed::EXTRA_DIR
                .get_file(format!(
                    "genesis-system-configs/{}/{}",
                    target_name, json_file_name
                ))
                .expect("Failed to find genesis system config file")
                .contents();
            let genesis_config: GenesisSystemConfig = serde_json::from_slice(genesis_config_data)
                .expect("Failed to deserialize genesis system config file");

            let id = chain_config.chain_id;
            chain_config.superchain = target_name.to_string();
            genesis_system_configs.insert(id, genesis_config);
            op_chains.insert(id, chain_config);
            addresses.insert(id, addrs);
            chain_ids.push(id);
        }

        match target_name {
            "mainnet" => {
                if let Ok(ci_mainnet_rpc) = std::env::var("CIRCLE_CI_MAINNET_RPC") {
                    entry_config.l1.public_rpc = ci_mainnet_rpc;
                }
            }
            "sepolia" | "sepolia-dev-0" => {
                if let Ok(ci_sepolia_rpc) = std::env::var("CIRCLE_CI_SEPOLIA_RPC") {
                    entry_config.l1.public_rpc = ci_sepolia_rpc;
                }
            }
            _ => {}
        }

        let impls = new_contract_implementations(target_name);
        implementations.insert(target_name.to_string(), impls);

        superchains.insert(
            target_name.to_string(),
            Superchain {
                superchain: target_name.to_string(),
                config: entry_config,
                chain_ids,
            },
        );
    }

    (
        superchains,
        op_chains,
        addresses,
        genesis_system_configs,
        implementations,
    )
}

fn new_contract_implementations(network: &str) -> ContractImplementations {
    let implementations_data = embed::IMPLEMENTATIONS_DIR
        .get_file("implementations.yaml")
        .expect("Failed to find implementations.yaml file")
        .contents();

    let mut global_implementations: ContractImplementations =
        serde_yaml::from_slice(implementations_data)
            .expect("Failed to deserialize implementations.yaml file");

    if network.is_empty() {
        return global_implementations;
    }

    let network_implementations_data = embed::IMPLEMENTATIONS_DIR
        .get_file(format!("networks/{}.yaml", network))
        .expect("Failed to find network implementations.yaml file")
        .contents();

    let network_implementations = serde_yaml::from_slice(network_implementations_data)
        .expect("Failed to deserialize network implementations.yaml file");

    global_implementations.merge(network_implementations);

    global_implementations
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_read_superchains() {
        let superchains = SUPERCHAINS.clone();
        assert_eq!(superchains.len(), 3);
        assert_eq!(superchains.get("mainnet").unwrap().superchain, "mainnet");
        assert_eq!(superchains.get("sepolia").unwrap().superchain, "sepolia");
    }

    #[test]
    fn test_read_chain_configs() {
        let superchains = SUPERCHAINS.clone();
        let mainnet = superchains.get("mainnet").unwrap();
        assert_eq!(mainnet.config.name, "Mainnet");
        assert_eq!(mainnet.chain_ids.len(), 8);
    }
}
