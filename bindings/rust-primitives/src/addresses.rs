//! Address Types

use alloy_primitives::Address;
use hashbrown::HashMap;

/// Map of chain IDs to their address lists.
pub type Addresses = HashMap<u64, AddressList>;

/// The set of network-specific contracts for a given chain.
#[derive(Debug, Clone, Hash, PartialEq, Eq, Default)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "PascalCase"))]
pub struct AddressList {
    /// The address manager
    pub address_manager: Address,
    /// L1 Cross Domain Messenger proxy address
    pub l1_cross_domain_messenger_proxy: Address,
    /// L1 ERC721 Bridge proxy address
    #[cfg_attr(feature = "serde", serde(alias = "L1ERC721BridgeProxy"))]
    pub l1_erc721_bridge_proxy: Address,
    /// L1 Standard Bridge proxy address
    pub l1_standard_bridge_proxy: Address,
    /// L2 Output Oracle Proxy address
    pub l2_output_oracle_proxy: Option<Address>,
    /// Optimism Mintable ERC20 Factory Proxy address
    #[cfg_attr(feature = "serde", serde(alias = "OptimismMintableERC20FactoryProxy"))]
    pub optimism_mintable_erc20_factory_proxy: Address,
    /// Optimism Portal Proxy address
    pub optimism_portal_proxy: Address,
    /// System Config Proxy address
    pub system_config_proxy: Address,
    /// Proxy Admin address
    pub proxy_admin: Address,

    // Fault Proof Contract Addresses
    /// Anchor State Registry Proxy address
    pub anchor_state_registry_proxy: Option<Address>,
    /// Delayed WETH Proxy address
    #[cfg_attr(feature = "serde", serde(alias = "DelayedWETHProxy"))]
    pub delayed_weth_proxy: Option<Address>,
    /// Dispute Game Factory Proxy address
    pub dispute_game_factory_proxy: Option<Address>,
    /// Fault Dispute Game Proxy address
    pub fault_dispute_game: Option<Address>,
    /// MIPS Proxy address
    #[cfg_attr(feature = "serde", serde(alias = "MIPS"))]
    pub mips: Option<Address>,
    /// Permissioned Dispute Game Proxy address
    pub permissioned_dispute_game: Option<Address>,
    /// Preimage Oracle Proxy address
    pub preimage_oracle: Option<Address>,
}
