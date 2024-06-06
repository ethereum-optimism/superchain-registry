//! Genesis types.

use crate::BlockID;
use crate::SystemConfig;
use alloy_primitives::{Address, Bytes, B256};
use hashbrown::HashMap;

/// Map of chain IDs to their chain's genesis system configurations.
pub type GenesisSystemConfigs = HashMap<u64, GenesisSystemConfig>;

/// Chain genesis information.
#[derive(Debug, Clone, Default)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
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
    #[cfg_attr(feature = "serde", serde(flatten))]
    pub system_config: Option<SystemConfig>,
}

/// Genesis system configuration.
#[derive(Debug, Clone, Default)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "camelCase"))]
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
