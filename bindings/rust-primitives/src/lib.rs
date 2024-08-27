#![doc = include_str!("../README.md")]
#![warn(missing_debug_implementations, missing_docs, rustdoc::all)]
#![deny(unused_must_use, rust_2018_idioms)]
#![cfg_attr(docsrs, feature(doc_cfg, doc_auto_cfg))]
#![cfg_attr(not(test), warn(unused_crate_dependencies))]
#![cfg_attr(not(feature = "std"), no_std)]

extern crate alloc;

/// Re-export the Genesis type from [alloy_genesis].
pub use alloy_genesis::Genesis;

pub mod superchain;
pub use superchain::{Superchain, SuperchainConfig, SuperchainL1Info, SuperchainLevel};

pub mod predeploys;
pub use predeploys::{
    BASE_FEE_VAULT, BEACON_BLOCK_ROOT, EAS, GAS_PRICE_ORACLE, GOVERNANCE_TOKEN, L1_BLOCK,
    L1_BLOCK_NUMBER, L1_FEE_VAULT, L2_CROSS_DOMAIN_MESSENGER, L2_ERC721_BRIDGE, L2_STANDARD_BRIDGE,
    L2_TO_L1_MESSAGE_PASSER, OPTIMISM_MINTABLE_ERC20_FACTORY, OPTIMISM_MINTABLE_ERC721_FACTORY,
    PROXY_ADMIN, SCHEMA_REGISTRY, SEQUENCER_FEE_VAULT, WETH9,
};

pub mod fee_params;
pub use fee_params::{
    BASE_SEPOLIA_BASE_FEE_PARAMS, BASE_SEPOLIA_CANYON_BASE_FEE_PARAMS,
    BASE_SEPOLIA_EIP1559_DEFAULT_ELASTICITY_MULTIPLIER, OP_BASE_FEE_PARAMS,
    OP_CANYON_BASE_FEE_PARAMS, OP_SEPOLIA_BASE_FEE_PARAMS, OP_SEPOLIA_CANYON_BASE_FEE_PARAMS,
    OP_SEPOLIA_EIP1559_BASE_FEE_MAX_CHANGE_DENOMINATOR_CANYON,
    OP_SEPOLIA_EIP1559_DEFAULT_BASE_FEE_MAX_CHANGE_DENOMINATOR,
    OP_SEPOLIA_EIP1559_DEFAULT_ELASTICITY_MULTIPLIER,
};

pub mod rollup_config;
pub use rollup_config::{
    load_op_stack_rollup_config, rollup_config_from_chain_id, RollupConfig, BASE_MAINNET_CONFIG,
    BASE_SEPOLIA_CONFIG, OP_MAINNET_CONFIG, OP_SEPOLIA_CONFIG,
};

pub mod chain_config;
pub use chain_config::{AltDAConfig, ChainConfig, HardForkConfiguration};

pub mod genesis;
pub use genesis::ChainGenesis;

pub mod block;
pub use block::BlockID;

pub mod system_config;
pub use system_config::SystemConfig;

pub mod addresses;
pub use addresses::AddressList;
