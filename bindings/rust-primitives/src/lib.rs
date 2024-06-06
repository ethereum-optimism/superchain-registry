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

mod chain_config;
pub use chain_config::{ChainConfig, HardForkConfiguration, OPChains, PlasmaConfig};

mod genesis;
pub use genesis::{ChainGenesis, GenesisSystemConfig, GenesisSystemConfigs};

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
