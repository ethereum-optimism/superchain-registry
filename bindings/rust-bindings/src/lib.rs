#![doc = include_str!("../README.md")]
#![warn(missing_debug_implementations, missing_docs, rustdoc::all)]
#![deny(unused_must_use, rust_2018_idioms)]
#![cfg_attr(docsrs, feature(doc_cfg, doc_auto_cfg))]
#![cfg_attr(not(feature = "std"), no_std)]

extern crate alloc;

/// Re-export commonly used types and traits.
pub use hashbrown::HashMap;
pub use superchain_primitives::*;

pub mod chain_list;
pub use chain_list::{Chain, ChainList};

pub mod superchain;
pub use superchain::Superchain;

pub use superchain_primitives::RollupConfigs;

lazy_static::lazy_static! {
    /// Private initializer that loads the superchain configurations.
    static ref _INIT: Superchain = Superchain::from_chain_list();

    /// Chain configurations exported from the registry
    pub static ref CHAINS: Vec<Chain> = _INIT.chains.clone();

    /// OP Chain configurations exported from the registry
    pub static ref OPCHAINS: OPChains = _INIT.op_chains.clone();

    /// Rollup configurations exported from the registry
    pub static ref ROLLUP_CONFIGS: RollupConfigs = _INIT.rollup_configs.clone();
}

#[cfg(test)]
mod tests {
    #[test]
    fn test_hardcoded_rollup_configs() {
        let test_cases = vec![
            (10, superchain_primitives::OP_MAINNET_CONFIG),
            (8453, superchain_primitives::BASE_MAINNET_CONFIG),
            (11155420, superchain_primitives::OP_SEPOLIA_CONFIG),
            (84532, superchain_primitives::BASE_SEPOLIA_CONFIG),
        ];

        for (chain_id, expected) in test_cases {
            let derived = super::ROLLUP_CONFIGS.get(&chain_id).unwrap();
            assert_eq!(expected, *derived);
        }
    }
}
