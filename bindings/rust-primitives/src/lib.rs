#![doc = include_str!("../README.md")]
#![warn(missing_debug_implementations, missing_docs, rustdoc::all)]
#![deny(unused_must_use, rust_2018_idioms)]
#![cfg_attr(docsrs, feature(doc_cfg, doc_auto_cfg))]
#![cfg_attr(not(feature = "std"), no_std)]

extern crate alloc;

/// Re-export the Genesis type from [alloy_genesis].
pub use alloy_genesis::Genesis;

mod superchain;
pub use superchain::{
    Superchain, SuperchainConfig, SuperchainL1Info, SuperchainLevel, Superchains,
};

mod rollup_config;
pub use rollup_config::{
    RollupConfig, OP_SEPOLIA_BASE_FEE_PARAMS, OP_SEPOLIA_CANYON_BASE_FEE_PARAMS,
    OP_SEPOLIA_EIP1559_BASE_FEE_MAX_CHANGE_DENOMINATOR_CANYON,
    OP_SEPOLIA_EIP1559_DEFAULT_BASE_FEE_MAX_CHANGE_DENOMINATOR,
    OP_SEPOLIA_EIP1559_DEFAULT_ELASTICITY_MULTIPLIER,
};

pub use rollup_config::{OP_MAINNET_CONFIG, OP_SEPOLIA_CONFIG, BASE_MAINNET_CONFIG, BASE_SEPOLIA_CONFIG};

mod chain_config;
pub use chain_config::{ChainConfig, HardForkConfiguration, OPChains, PlasmaConfig};

mod genesis;
pub use genesis::{ChainGenesis, GenesisSystemConfigs};

mod block;
pub use block::BlockID;

mod system_config;
pub use system_config::SystemConfig;

mod addresses;
pub use addresses::{AddressList, Addresses};

mod contracts;
pub use contracts::{AddressSet, ContractImplementations, Implementations};

/// Validates if a file is a configuration file.
pub fn is_config_file(name: &str) -> bool {
    name.ends_with(".yaml") && name != "superchain.yaml" && name != "semver.yaml"
}
