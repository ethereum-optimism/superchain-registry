#![doc = include_str!("../README.md")]
#![warn(missing_debug_implementations, missing_docs, rustdoc::all)]
#![deny(unused_must_use, rust_2018_idioms)]
#![cfg_attr(docsrs, feature(doc_cfg, doc_auto_cfg))]

/// Module holding the superchain configuration definition types.
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
    use crate::SUPERCHAINS;

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
