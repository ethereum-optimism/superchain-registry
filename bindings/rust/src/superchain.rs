use std::collections::HashMap;

use serde::{Deserialize, Serialize};
use serde_repr::{Deserialize_repr, Serialize_repr};

pub type Superchains = HashMap<String, Superchain>;
pub type OPChains = HashMap<u64, ChainConfig>;
pub type Addresses = HashMap<u64, AddressList>;
pub type GenesisSystemConfigs = HashMap<u64, GenesisSystemConfig>;
pub type Implementations = HashMap<String, ContractImplementations>;

// TODO: should we use alloy for these?
pub type Address = String;
pub type Hash = String;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Superchain {
    /// Superchain configuration.
    pub config: SuperchainConfig,
    /// Chain IDs of chains that are part of this superchain.
    pub chain_ids: Vec<u64>,
    /// Superchain identifier, without capitalization or display changes.
    pub superchain: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SuperchainConfig {
    pub name: String,
    pub l1: SuperchainL1Info,

    pub protocol_versions_addr: Option<Address>,
    pub superchain_config_addr: Option<Address>,

    /// Hardfork Configuration. These values may be overridden by individual chains.
    #[serde(flatten)]
    pub hardfork_defaults: HardForkConfiguration,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SuperchainL1Info {
    pub chain_id: u64,
    pub public_rpc: String,
    pub explorer: String,
}

#[derive(Debug, Clone, Serialize_repr, Deserialize_repr)]
#[repr(u8)]
pub enum SuperchainLevel {
    Frontier = 1,
    Standard = 2,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChainConfig {
    pub name: String,
    pub chain_id: u64,
    pub public_rpc: String,
    pub sequencer_rpc: String,
    pub explorer: String,
    pub superchain_level: SuperchainLevel,
    /// If superchain_time is set, hardforks times after superchain_time
    /// will be inherited from the superchain-wide config.
    pub superchain_time: Option<u64>,
    pub batch_inbox_addr: Address,
    pub genesis: ChainGenesis,
    /// Superchain is a simple string to identify the superchain.
    /// This is implied by directory structure, and not encoded in the config file itself.
    #[serde(skip)]
    pub superchain: String,
    /// Chain is a simple string to identify the chain, within its superchain context.
    /// This matches the resource filename, it is not encoded in the config file itself.
    #[serde(skip)]
    pub chain: String,
    #[serde(flatten)]
    pub hardfork_configuration: HardForkConfiguration,
    /// Optional feature
    pub plasma: Option<PlasmaConfig>,
}

impl ChainConfig {
    /// Set missing hardfork configurations to the defaults, if the chain has
    /// a superchain_time set. Defaults are only used if the chain's hardfork
    /// activated after the superchain_time.
    pub(crate) fn set_missing_fork_configs(&mut self, defaults: &HardForkConfiguration) {
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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BlockID {
    pub hash: Hash,
    pub number: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChainGenesis {
    pub l1: BlockID,
    pub l2: BlockID,
    pub l2_time: u64,
    pub extra_data: Option<Vec<u8>>,
    #[serde(flatten)]
    pub system_config: Option<SystemConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SystemConfig {
    pub batcher_addr: String,
    pub overhead: String,
    pub scalar: String,
    pub gas_limit: u64,
    pub base_fee_scalar: u64,
    pub blob_base_fee_scalar: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlasmaConfig {
    pub da_challenge_address: Option<Address>,
    pub da_challenge_window: Option<u64>,
    pub da_resolve_window: Option<u64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HardForkConfiguration {
    pub canyon_time: Option<u64>,
    pub delta_time: Option<u64>,
    pub ecotone_time: Option<u64>,
    pub fjord_time: Option<u64>,
}

/// The set of network-specific contracts for a given chain.
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
pub struct AddressList {
    pub address_manager: Address,
    pub l1_cross_domain_messenger_proxy: Address,
    #[serde(alias = "L1ERC721BridgeProxy")]
    pub l1_erc721_bridge_proxy: Address,
    pub l1_standard_bridge_proxy: Address,
    pub l2_output_oracle_proxy: Option<Address>,
    #[serde(alias = "OptimismMintableERC20FactoryProxy")]
    pub optimism_mintable_erc20_factory_proxy: Address,
    pub optimism_portal_proxy: Address,
    pub system_config_proxy: Address,
    pub proxy_admin: Address,

    // Fault proof contracts:
    pub anchor_state_registry_proxy: Option<Address>,
    #[serde(alias = "DelayedWETHProxy")]
    pub delayed_weth_proxy: Option<Address>,
    pub dispute_game_factory_proxy: Option<Address>,
    pub fault_dispute_game: Option<Address>,
    #[serde(alias = "MIPS")]
    pub mips: Option<Address>,
    pub permissioned_dispute_game: Option<Address>,
    pub preimage_oracle: Option<Address>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct GenesisSystemConfig {
    pub batcher_addr: Address,
    pub overhead: Hash,
    pub scalar: Hash,
    pub gas_limit: u64,
}

/// A set of addresses for a given contract. The key is the semver version.
pub type AddressSet = HashMap<String, Address>;

#[derive(Debug, Default, Clone, Serialize, Deserialize)]
#[serde(default)]
pub struct ContractImplementations {
    pub l1_cross_domain_messenger: AddressSet,
    pub l1_erc721_bridge: AddressSet,
    pub l1_standard_bridge: AddressSet,
    pub l2_output_oracle: AddressSet,
    pub optimism_mintable_erc20_factory: AddressSet,
    pub optimism_portal: AddressSet,
    pub system_config: AddressSet,

    // Fault Proof contracts:
    pub anchor_state_registry: AddressSet,
    pub delayed_weth: AddressSet,
    pub dispute_game_factory: AddressSet,
    pub fault_dispute_game: AddressSet,
    pub mips: AddressSet,
    pub permissioned_dispute_game: AddressSet,
    pub preimage_oracle: AddressSet,
}

impl ContractImplementations {
    pub(crate) fn merge(&mut self, other: Self) {
        self.l1_cross_domain_messenger
            .extend(other.l1_cross_domain_messenger);
        self.l1_erc721_bridge.extend(other.l1_erc721_bridge);
        self.l1_standard_bridge.extend(other.l1_standard_bridge);
        self.l2_output_oracle.extend(other.l2_output_oracle);
        self.optimism_mintable_erc20_factory
            .extend(other.optimism_mintable_erc20_factory);
        self.optimism_portal.extend(other.optimism_portal);
        self.system_config.extend(other.system_config);

        // Fault Proof contracts:
        self.anchor_state_registry
            .extend(other.anchor_state_registry);
        self.delayed_weth.extend(other.delayed_weth);
        self.dispute_game_factory.extend(other.dispute_game_factory);
        self.fault_dispute_game.extend(other.fault_dispute_game);
        self.mips.extend(other.mips);
        self.permissioned_dispute_game
            .extend(other.permissioned_dispute_game);
        self.preimage_oracle.extend(other.preimage_oracle);
    }
}

pub(crate) fn is_config_file(name: &str) -> bool {
    name.ends_with(".yaml") && name != "superchain.yaml" && name != "semver.yaml"
}
