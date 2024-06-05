#![doc = include_str!("../README.md")]
#![warn(missing_debug_implementations, missing_docs, rustdoc::all)]
#![deny(unused_must_use, rust_2018_idioms)]
#![cfg_attr(docsrs, feature(doc_cfg, doc_auto_cfg))]
#![no_std]

extern crate alloc;

/// Module holding the superchain configuration type definitions.
pub mod superchain;
use superchain::{Addresses, GenesisSystemConfigs, Implementations, OPChains, Superchains};

/// Module responsible for representing the embedded filesystem
/// that contains the superchain configurations for static access.
mod embed;

/// Module responsible for initializing the superchain configurations
/// from the embedded filesystem.
mod init;

lazy_static::lazy_static! {
    /// Private initializer that runs once to load the superchain configurations.
    static ref _INIT: init::InitTuple = init::load_embedded_configs();

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

#[cfg(test)]
mod tests {
    use alloy_primitives::{address, b256};

    use crate::{superchain::SuperchainLevel, OPCHAINS, SUPERCHAINS};

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

        let base_chain_id = 8453;
        let opchains = OPCHAINS.clone();
        let base = opchains.get(&base_chain_id).unwrap();
        assert_eq!(base.chain_id, base_chain_id);
        assert_eq!(base.name, "Base");
        assert_eq!(base.superchain, "mainnet");
        assert_eq!(base.chain, "base");
        assert_eq!(base.public_rpc, "https://mainnet.base.org");
        assert_eq!(base.sequencer_rpc, "https://mainnet-sequencer.base.org");
        assert_eq!(base.explorer, "https://explorer.base.org");

        assert!(matches!(base.superchain_level, SuperchainLevel::Standard));
        assert_eq!(base.superchain_time, Some(0));
        assert_eq!(
            base.batch_inbox_addr,
            address!("ff00000000000000000000000000000000008453")
        );

        assert_eq!(base.genesis.l1.number, 17481768);
        assert_eq!(
            base.genesis.l1.hash,
            b256!("5c13d307623a926cd31415036c8b7fa14572f9dac64528e857a470511fc30771")
        );
        assert_eq!(base.genesis.l2.number, 0);
        assert_eq!(
            base.genesis.l2.hash,
            b256!("f712aa9241cc24369b143cf6dce85f0902a9731e70d66818a3a5845b296c73dd")
        );
        assert_eq!(base.genesis.l2_time, 1686789347);
        assert_eq!(base.genesis.extra_data, None);
        assert!(base.genesis.system_config.is_none());

        assert_eq!(base.hardfork_configuration.canyon_time, None);
        assert_eq!(base.hardfork_configuration.delta_time, None);
        assert_eq!(base.hardfork_configuration.ecotone_time, None);
        assert_eq!(base.hardfork_configuration.fjord_time, None);
    }
}
