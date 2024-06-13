//! Contracts

use alloc::string::String;
use alloy_primitives::Address;
use hashbrown::HashMap;

/// Map of superchain names to their implementation contract semvers.
pub type Implementations = HashMap<String, ContractImplementations>;

/// A set of addresses for a given contract. The key is the semver version.
pub type AddressSet = HashMap<String, Address>;

/// Contract Implementations
#[derive(Debug, Default, Clone, Eq, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(default))]
pub struct ContractImplementations {
    /// L1 Cross Domain Messenger
    pub l1_cross_domain_messenger: AddressSet,
    /// L1 ERC721 Bridge
    pub l1_erc721_bridge: AddressSet,
    /// L1 Standard Bridge
    pub l1_standard_bridge: AddressSet,
    /// L2 Output Oracle
    pub l2_output_oracle: AddressSet,
    /// Optimism Mintable ERC20 Factory
    pub optimism_mintable_erc20_factory: AddressSet,
    /// Optimism Portal
    pub optimism_portal: AddressSet,
    /// System Config
    pub system_config: AddressSet,

    // Fault Proof Contracts
    /// Anchor State Registry
    pub anchor_state_registry: AddressSet,
    /// Delayed WETH
    pub delayed_weth: AddressSet,
    /// Dispute Game Factory
    pub dispute_game_factory: AddressSet,
    /// Fault Dispute Game
    pub fault_dispute_game: AddressSet,
    /// MIPS
    pub mips: AddressSet,
    /// Permissioned Dispute Game
    pub permissioned_dispute_game: AddressSet,
    /// Preimage Oracle
    pub preimage_oracle: AddressSet,
}

impl ContractImplementations {
    /// Merges two sets of contract implementations.
    pub fn merge(&mut self, other: Self) {
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
