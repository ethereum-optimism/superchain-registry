#![doc = include_str!("../README.md")]
#![warn(missing_debug_implementations, missing_docs, rustdoc::all)]
#![deny(unused_must_use, rust_2018_idioms)]
#![cfg_attr(docsrs, feature(doc_cfg, doc_auto_cfg))]
#![cfg_attr(not(feature = "std"), no_std)]

extern crate alloc;

/// Module responsible for initializing the superchain configurations
/// from the embedded filesystem.
mod init;

/// Re-export superchain primitives.
pub use superchain_primitives::{
    Addresses, GenesisSystemConfigs, Implementations, OPChains, Superchains,
};

use include_dir::{include_dir, Dir};

/// Directory containing the configuration files for the superchain.
pub(crate) static CONFIGS_DIR: Dir<'_> = include_dir!("$SUPERCHAIN_CONFIGS");

lazy_static::lazy_static! {
    /// Private initializer that runs once to load the superchain configurations.
    static ref _INIT: init::InitTuple = init::load_embedded_configs();

    /// Superchain configurations exported from the registry
    pub static ref SUPERCHAINS: Superchains = _INIT.0.clone();

    /// OPChain configurations exported from the registry
    pub static ref OPCHAINS: OPChains = _INIT.1.clone();
}

#[cfg(test)]
mod tests {
    use crate::{OPCHAINS, SUPERCHAINS};
    use alloy_primitives::{address, b256, uint};
    use superchain_primitives::{
        AddressList, SuperchainLevel, SystemConfig, BASE_MAINNET_CONFIG, OP_MAINNET_CONFIG,
    };

    #[test]
    fn test_read_superchains() {
        let superchains = SUPERCHAINS.clone();
        assert_eq!(superchains.len(), 3);
        assert_eq!(superchains.get("mainnet").unwrap().superchain, "mainnet");
        assert_eq!(superchains.get("sepolia").unwrap().superchain, "sepolia");
    }

    #[test]
    fn test_base_mainnet_genesis() {
        let base = BASE_MAINNET_CONFIG;
        let derived = OPCHAINS.get(&8453).unwrap();
        assert_eq!(base.genesis, derived.genesis);
    }

    #[test]
    fn test_op_mainnet_genesis() {
        let op = OP_MAINNET_CONFIG;
        let derived = OPCHAINS.get(&10).unwrap();
        assert_eq!(op.genesis, derived.genesis);
    }

    #[test]
    fn test_read_chain_configs() {
        let mainnet = SUPERCHAINS.get("mainnet").unwrap();
        assert_eq!(mainnet.config.name, "Mainnet");

        let base_chain_id = 8453;
        let base = OPCHAINS.get(&base_chain_id).unwrap();
        assert_eq!(base.chain_id, base_chain_id);
        assert_eq!(base.name, "Base");
        assert_eq!(base.superchain, "mainnet");
        assert_eq!(base.chain, "base");
        assert_eq!(base.public_rpc, "https://mainnet.base.org");
        assert_eq!(base.sequencer_rpc, "https://mainnet-sequencer.base.org");
        assert_eq!(base.explorer, "https://explorer.base.org");

        assert!(matches!(base.superchain_level, SuperchainLevel::Frontier));
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

        let sys_config = SystemConfig {
            batcher_address: address!("5050f69a9786f081509234f1a7f4684b5e5b76c9"),
            overhead: uint!(0xbc_U256),
            scalar: uint!(0xa6fe0_U256),
            gas_limit: 30_000_000_u64,
            base_fee_scalar: None,
            blob_base_fee_scalar: None,
        };
        assert_eq!(base.genesis.system_config, Some(sys_config));

        assert_eq!(base.hardfork_configuration.canyon_time, Some(1704992401));
        assert_eq!(base.hardfork_configuration.delta_time, Some(1708560000));
        assert_eq!(base.hardfork_configuration.ecotone_time, Some(1710374401));
        assert_eq!(base.hardfork_configuration.fjord_time, Some(1720627201));

        let expected = AddressList {
            address_manager: address!("8efb6b5c4767b09dc9aa6af4eaa89f749522bae2"),
            l1_cross_domain_messenger_proxy: address!("866e82a600a1414e583f7f13623f1ac5d58b0afa"),
            l1_erc721_bridge_proxy: address!("608d94945a64503e642e6370ec598e519a2c1e53"),
            l1_standard_bridge_proxy: address!("3154cf16ccdb4c6d922629664174b904d80f2c35"),
            l2_output_oracle_proxy: Some(address!("56315b90c40730925ec5485cf004d835058518a0")),
            optimism_mintable_erc20_factory_proxy: address!(
                "05cc379ebd9b30bba19c6fa282ab29218ec61d84"
            ),
            optimism_portal_proxy: address!("49048044d57e1c92a77f79988d21fa8faf74e97e"),
            system_config_proxy: address!("73a79fab69143498ed3712e519a88a918e1f4072"),
            proxy_admin: address!("0475cbcaebd9ce8afa5025828d5b98dfb67e059e"),
            anchor_state_registry_proxy: None,
            delayed_weth_proxy: None,
            dispute_game_factory_proxy: None,
            fault_dispute_game: None,
            mips: None,
            permissioned_dispute_game: None,
            preimage_oracle: None,
        };

        assert_eq!(base.addresses, Some(expected));
    }
}
