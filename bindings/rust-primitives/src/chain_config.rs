//! Chain Config Types

use crate::ChainGenesis;
use crate::SuperchainLevel;
use alloc::string::String;
use alloy_primitives::Address;
use hashbrown::HashMap;

/// Map of OPChain IDs to their [ChainConfig].
pub type OPChains = HashMap<u64, ChainConfig>;

/// Plasma configuration.
#[derive(Debug, Clone, Default, Eq, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct PlasmaConfig {
    /// Plasma DA challenge address
    pub da_challenge_address: Option<Address>,
    /// Plasma DA challenge window time (in seconds)
    pub da_challenge_window: Option<u64>,
    /// Plasma DA resolution window time (in seconds)
    pub da_resolve_window: Option<u64>,
}

/// Hardfork configuration.
#[derive(Debug, Clone, Default, Eq, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct HardForkConfiguration {
    /// Canyon hardfork activation time
    pub canyon_time: Option<u64>,
    /// Delta hardfork activation time
    pub delta_time: Option<u64>,
    /// Ecotone hardfork activation time
    pub ecotone_time: Option<u64>,
    /// Fjord hardfork activation time
    pub fjord_time: Option<u64>,
}

/// A chain configuration.
#[derive(Debug, Clone, Default, Eq, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct ChainConfig {
    /// Chain name (e.g. "Base")
    pub name: String,
    /// Chain ID
    pub chain_id: u64,
    /// Chain public RPC endpoint
    pub public_rpc: String,
    /// Chain sequencer RPC endpoint
    pub sequencer_rpc: String,
    /// Chain explorer HTTP endpoint
    pub explorer: String,
    /// Level of integration with the superchain.
    pub superchain_level: SuperchainLevel,
    /// Time of opt-in to the Superchain.
    /// If superchain_time is set, hardforks times after superchain_time
    /// will be inherited from the superchain-wide config.
    pub superchain_time: Option<u64>,
    /// Chain-specific batch inbox address
    pub batch_inbox_addr: Address,
    /// Chain-specific genesis information
    pub genesis: ChainGenesis,
    /// Superchain is a simple string to identify the superchain.
    /// This is implied by directory structure, and not encoded in the config file itself.
    #[cfg_attr(feature = "serde", serde(skip))]
    pub superchain: String,
    /// Chain is a simple string to identify the chain, within its superchain context.
    /// This matches the resource filename, it is not encoded in the config file itself.
    #[cfg_attr(feature = "serde", serde(skip))]
    pub chain: String,
    #[cfg_attr(feature = "serde", serde(flatten))]
    /// Hardfork Configuration. These values may override the superchain-wide defaults.
    pub hardfork_configuration: HardForkConfiguration,
    /// Optional Plasma DA feature
    pub plasma: Option<PlasmaConfig>,
}

impl ChainConfig {
    /// Set missing hardfork configurations to the defaults, if the chain has
    /// a superchain_time set. Defaults are only used if the chain's hardfork
    /// activated after the superchain_time.
    pub fn set_missing_fork_configs(&mut self, defaults: &HardForkConfiguration) {
        let Some(super_time) = self.superchain_time else {
            return;
        };
        let cfg = &mut self.hardfork_configuration;

        if cfg.canyon_time.is_some_and(|t| t > super_time) {
            cfg.canyon_time = defaults.canyon_time;
        }
        if cfg.delta_time.is_some_and(|t| t > super_time) {
            cfg.delta_time = defaults.delta_time;
        }
        if cfg.ecotone_time.is_some_and(|t| t > super_time) {
            cfg.ecotone_time = defaults.ecotone_time;
        }
        if cfg.fjord_time.is_some_and(|t| t > super_time) {
            cfg.fjord_time = defaults.fjord_time;
        }
    }
}
