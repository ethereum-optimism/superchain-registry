use alloc::{string::String, vec::Vec};
use alloy_primitives::{Address, Bytes, B256};
use hashbrown::HashMap;
use serde::{Deserialize, Serialize};
use serde_repr::{Deserialize_repr, Serialize_repr};

pub use alloy_genesis::Genesis;

/// Map of superchain names to their configurations.
pub type Superchains = HashMap<String, Superchain>;

/// Map of OPChain IDs to their configurations.
pub type OPChains = HashMap<u64, ChainConfig>;

/// Map of chain IDs to their address lists.
pub type Addresses = HashMap<u64, AddressList>;

/// Map of chain IDs to their chain's genesis system configurations.
pub type GenesisSystemConfigs = HashMap<u64, GenesisSystemConfig>;

/// Map of superchain names to their implementation contract semvers.
pub type Implementations = HashMap<String, ContractImplementations>;

/// A superchain configuration.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Superchain {
    /// Superchain configuration file contents.
    pub config: SuperchainConfig,
    /// Chain IDs of chains that are part of this superchain.
    pub chain_ids: Vec<u64>,
    /// Superchain identifier, without capitalization or display changes.
    pub superchain: String,
}

/// A superchain configuration file format
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SuperchainConfig {
    /// Superchain name (e.g. "Mainnet")
    pub name: String,
    /// Superchain L1 anchor information
    pub l1: SuperchainL1Info,
    /// Optional addresses for the superchain-wide default protocol versions contract.
    pub protocol_versions_addr: Option<Address>,
    /// Optional address for the superchain-wide default superchain config contract.
    pub superchain_config_addr: Option<Address>,
    /// Hardfork Configuration. These values may be overridden by individual chains.
    #[serde(flatten)]
    pub hardfork_defaults: HardForkConfiguration,
}

/// Superchain L1 anchor information
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SuperchainL1Info {
    /// L1 chain ID
    pub chain_id: u64,
    /// L1 chain public RPC endpoint
    pub public_rpc: String,
    /// L1 chain explorer RPC endpoint
    pub explorer: String,
}

/// Level of integration with the superchain.
#[derive(Debug, Clone, Default, Serialize_repr, Deserialize_repr)]
#[repr(u8)]
pub enum SuperchainLevel {
    /// Frontier chains are chains with customizations beyond the
    /// standard OP Stack configuration and are considered "advanced".
    Frontier = 1,
    /// Standard chains don't have any customizations beyond the
    /// standard OP Stack configuration and are considered "vanilla".
    #[default]
    Standard = 2,
}

/// A chain configuration.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
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
    #[serde(skip)]
    pub superchain: String,
    /// Chain is a simple string to identify the chain, within its superchain context.
    /// This matches the resource filename, it is not encoded in the config file itself.
    #[serde(skip)]
    pub chain: String,
    #[serde(flatten)]
    /// Hardfork Configuration. These values may override the superchain-wide defaults.
    pub hardfork_configuration: HardForkConfiguration,
    /// Optional Plasma DA feature
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

/// Block identifier.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct BlockID {
    /// Block hash
    pub hash: B256,
    /// Block number
    pub number: u64,
}

/// Chain genesis information.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ChainGenesis {
    /// L1 genesis block
    pub l1: BlockID,
    /// L2 genesis block
    pub l2: BlockID,
    /// Timestamp of the L2 genesis block
    pub l2_time: u64,
    /// Extra data for the genesis block
    pub extra_data: Option<Bytes>,
    /// Optional System configuration
    #[serde(flatten)]
    pub system_config: Option<SystemConfig>,
}

/// System configuration.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SystemConfig {
    /// Batcher address
    pub batcher_addr: Address,
    /// Fee overhead value
    pub overhead: String,
    /// Fee scalar value
    pub scalar: String,
    /// Gas limit value
    pub gas_limit: u64,
    /// Base fee scalar value
    pub base_fee_scalar: u64,
    /// Blob base fee scalar value
    pub blob_base_fee_scalar: u64,
}

/// Plasma configuration.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct PlasmaConfig {
    /// Plasma DA challenge address
    pub da_challenge_address: Option<Address>,
    /// Plasma DA challenge window time (in seconds)
    pub da_challenge_window: Option<u64>,
    /// Plasma DA resolution window time (in seconds)
    pub da_resolve_window: Option<u64>,
}

/// Hardfork configuration.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
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

/// The set of network-specific contracts for a given chain.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "PascalCase")]
#[allow(missing_docs)]
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

/// Genesis system configuration.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct GenesisSystemConfig {
    /// Batcher address
    pub batcher_addr: Address,
    /// Fee overhead value
    pub overhead: B256,
    /// Fee scalar value
    pub scalar: B256,
    /// Gas limit value
    pub gas_limit: u64,
}

/// A set of addresses for a given contract. The key is the semver version.
pub type AddressSet = HashMap<String, Address>;

#[derive(Debug, Default, Clone, Serialize, Deserialize)]
#[serde(default)]
#[allow(missing_docs)]
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
