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

pub mod predeploys;
pub use predeploys::{
    BASE_FEE_VAULT, BEACON_BLOCK_ROOT, EAS, GAS_PRICE_ORACLE, GOVERNANCE_TOKEN, L1_BLOCK,
    L1_BLOCK_NUMBER, L1_FEE_VAULT, L2_CROSS_DOMAIN_MESSENGER, L2_ERC721_BRIDGE, L2_STANDARD_BRIDGE,
    L2_TO_L1_MESSAGE_PASSER, OPTIMISM_MINTABLE_ERC20_FACTORY, OPTIMISM_MINTABLE_ERC721_FACTORY,
    PROXY_ADMIN, SCHEMA_REGISTRY, SEQUENCER_FEE_VAULT, WETH9,
};

mod rollup_config;
pub use rollup_config::{
    RollupConfig, BASE_SEPOLIA_BASE_FEE_PARAMS, BASE_SEPOLIA_CANYON_BASE_FEE_PARAMS,
    BASE_SEPOLIA_EIP1559_DEFAULT_ELASTICITY_MULTIPLIER, OP_BASE_FEE_PARAMS,
    OP_CANYON_BASE_FEE_PARAMS, OP_SEPOLIA_BASE_FEE_PARAMS, OP_SEPOLIA_CANYON_BASE_FEE_PARAMS,
    OP_SEPOLIA_EIP1559_BASE_FEE_MAX_CHANGE_DENOMINATOR_CANYON,
    OP_SEPOLIA_EIP1559_DEFAULT_BASE_FEE_MAX_CHANGE_DENOMINATOR,
    OP_SEPOLIA_EIP1559_DEFAULT_ELASTICITY_MULTIPLIER,
};

pub use rollup_config::{
    BASE_MAINNET_CONFIG, BASE_SEPOLIA_CONFIG, OP_MAINNET_CONFIG, OP_SEPOLIA_CONFIG,
};

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
