//! Chain Config Types

use crate::AddressList;
use crate::ChainGenesis;
use crate::SuperchainLevel;
use alloc::string::String;
use alloy_primitives::Address;

/// AltDA configuration.
#[derive(Debug, Clone, Default, Hash, Eq, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct AltDAConfig {
    /// AltDA challenge address
    pub da_challenge_address: Option<Address>,
    /// AltDA challenge window time (in seconds)
    pub da_challenge_window: Option<u64>,
    /// AltDA resolution window time (in seconds)
    pub da_resolve_window: Option<u64>,
}

/// Hardfork configuration.
#[derive(Debug, Clone, Default, Hash, Eq, PartialEq)]
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
    /// Granite hardfork activation time
    pub granite_time: Option<u64>,
    /// Holocene hardfork activation time
    pub holocene_time: Option<u64>,
}

/// A chain configuration.
#[derive(Debug, Clone, Default, Hash, Eq, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct ChainConfig {
    /// Chain name (e.g. "Base")
    pub name: String,
    /// Chain ID
    pub chain_id: u64,
    /// L1 chain ID
    #[cfg_attr(feature = "serde", serde(skip))]
    pub l1_chain_id: u64,
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
    /// Hardfork Configuration. These values may override the superchain-wide defaults.
    #[cfg_attr(feature = "serde", serde(flatten))]
    pub hardfork_configuration: HardForkConfiguration,
    /// Optional AltDA feature
    pub alt_da: Option<AltDAConfig>,
    /// Addresses
    pub addresses: Option<AddressList>,
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

        if cfg.canyon_time.is_none() && defaults.canyon_time.is_some_and(|t| t > super_time) {
            cfg.canyon_time = defaults.canyon_time;
        }
        if cfg.delta_time.is_none() && defaults.delta_time.is_some_and(|t| t > super_time) {
            cfg.delta_time = defaults.delta_time;
        }
        if cfg.ecotone_time.is_none() && defaults.ecotone_time.is_some_and(|t| t > super_time) {
            cfg.ecotone_time = defaults.ecotone_time;
        }
        if cfg.fjord_time.is_none() && defaults.fjord_time.is_some_and(|t| t > super_time) {
            cfg.fjord_time = defaults.fjord_time;
        }
        if cfg.granite_time.is_none() && defaults.granite_time.is_some_and(|t| t > super_time) {
            cfg.granite_time = defaults.granite_time;
        }
        if cfg.holocene_time.is_none() && defaults.holocene_time.is_some_and(|t| t > super_time) {
            cfg.holocene_time = defaults.holocene_time;
        }
    }
}
