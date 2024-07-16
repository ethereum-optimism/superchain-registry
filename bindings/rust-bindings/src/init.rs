//! Initializers for the Superchain Registry.

use crate::CONFIGS_DIR;
use alloc::{format, vec::Vec};
use hashbrown::HashMap;
use superchain_primitives::{
    is_config_file, ChainConfig, OPChains, Superchain, SuperchainConfig, Superchains,
};

/// Tuple type holding the various initializers.
pub(crate) type InitTuple = (Superchains, OPChains);

/// Initialize the superchain configurations from the embedded filesystem.
pub(crate) fn load_embedded_configs() -> InitTuple {
    let mut superchains = HashMap::new();
    let mut op_chains = HashMap::new();

    for target_dir in CONFIGS_DIR.dirs() {
        let target_name = target_dir.path().file_name().unwrap().to_str().unwrap();
        let target_data = target_dir
            .get_file(format!("{}/superchain.toml", target_name))
            .expect("Failed to find superchain.toml config file")
            .contents_utf8()
            .expect("Failed to parse superchain.toml as utf8 string");
        let mut entry_config: SuperchainConfig =
            toml::from_str(target_data).expect("Failed to deserialize superchain.toml config file");

        let mut chain_ids = Vec::new();
        for chain in target_dir.entries() {
            let chain_name = chain.path().file_name().unwrap().to_str().unwrap();
            if !is_config_file(chain_name) {
                continue;
            }

            let chain_data = chain.as_file().unwrap().contents_utf8().unwrap();
            let mut chain_config: ChainConfig =
                toml::from_str(chain_data).expect("Failed to deserialize chain config file");
            chain_config.chain = chain_name.trim_end_matches(".toml").to_string();
            chain_config.set_missing_fork_configs(&entry_config.hardfork_defaults);

            chain_config.superchain = target_name.to_string();
            chain_ids.push(chain_config.chain_id);
            op_chains.insert(chain_config.chain_id, chain_config);
        }

        #[cfg(feature = "std")]
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

        superchains.insert(
            target_name.to_string(),
            Superchain {
                superchain: target_name.to_string(),
                config: entry_config,
                chain_ids,
            },
        );
    }

    (superchains, op_chains)
}
