//! Contains all [predeploy contract addresses][predeploys].
//!
//! Predeploys are smart contracts on the OP Stack that exist at predetermined addresses
//! at the genesis state. They are similar to precompiles but run directly in the EVM
//! instead of running native code outside of the EVM.
//!
//! Predeploys are used instead of precompiles to make it easier for multiclient
//! implementations and allow for more integration with hardhat/foundry network forking.
//!
//! Predeploy addresses exist in the 1-byte namespace `0x42000000000000000000000000000000000000xx`.
//! Proxies are set at each possible predeploy address except for the GovernanceToken and the ProxyAdmin.
//!
//! [predeploys]: https://specs.optimism.io/protocol/predeploys.html

use alloy_primitives::{address, Address};

/// Legacy Message Passer Predeploy.
#[deprecated]
pub const LEGACY_MESSAGE_PASSER: Address = address!("4200000000000000000000000000000000000000");

/// Deployer Whitelist Predeploy.
#[deprecated]
pub const DEPLOYER_WHITELIST: Address = address!("4200000000000000000000000000000000000002");

/// Legacy ERC20ETH Predeploy.
#[deprecated]
pub const LEGACY_ERC20_ETH: Address = address!("DeadDeAddeAddEAddeadDEaDDEAdDeaDDeAD0000");

/// WETH9 Predeploy.
pub const WETH9: Address = address!("4200000000000000000000000000000000000006");

/// L2 Cross Domain Messenger Predeploy.
pub const L2_CROSS_DOMAIN_MESSENGER: Address = address!("4200000000000000000000000000000000000007");

/// L2 Standard Bridge Predeploy.
pub const L2_STANDARD_BRIDGE: Address = address!("4200000000000000000000000000000000000010");

/// Sequencer Fee Vault Predeploy.
pub const SEQUENCER_FEE_VAULT: Address = address!("4200000000000000000000000000000000000011");

/// Optimism Mintable ERC20 Factory Predeploy.
pub const OPTIMISM_MINTABLE_ERC20_FACTORY: Address =
    address!("4200000000000000000000000000000000000012");

/// L1 Block Number Predeploy.
pub const L1_BLOCK_NUMBER: Address = address!("4200000000000000000000000000000000000013");

/// Gas Price Oracle Predeploy.
pub const GAS_PRICE_ORACLE: Address = address!("420000000000000000000000000000000000000F");

/// Governance Token Predeploy.
pub const GOVERNANCE_TOKEN: Address = address!("4200000000000000000000000000000000000042");

/// L1 Block Predeploy.
pub const L1_BLOCK: Address = address!("4200000000000000000000000000000000000015");

/// L2 To L1 Message Passer Predeploy.
pub const L2_TO_L1_MESSAGE_PASSER: Address = address!("4200000000000000000000000000000000000016");

/// L2 ERC721 Bridge Predeploy.
pub const L2_ERC721_BRIDGE: Address = address!("4200000000000000000000000000000000000014");

/// Optimism Mintable ERC721 Factory Predeploy.
pub const OPTIMISM_MINTABLE_ERC721_FACTORY: Address =
    address!("4200000000000000000000000000000000000017");

/// Proxy Admin Predeploy.
pub const PROXY_ADMIN: Address = address!("4200000000000000000000000000000000000018");

/// Base Fee Vault Predeploy.
pub const BASE_FEE_VAULT: Address = address!("4200000000000000000000000000000000000019");

/// L1 Fee Vault Predeploy.
pub const L1_FEE_VAULT: Address = address!("420000000000000000000000000000000000001a");

/// Schema Registry Predeploy.
pub const SCHEMA_REGISTRY: Address = address!("4200000000000000000000000000000000000020");

/// EAS Predeploy.
pub const EAS: Address = address!("4200000000000000000000000000000000000021");

/// Beacon Block Root Predeploy.
pub const BEACON_BLOCK_ROOT: Address = address!("000F3df6D732807Ef1319fB7B8bB8522d0Beac02");
