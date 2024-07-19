//! Genesis types.

use crate::BlockID;
use crate::SystemConfig;
use alloy_primitives::Bytes;

/// Chain genesis information.
#[derive(Debug, Clone, Default, Hash, Eq, PartialEq)]
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
    pub system_config: Option<SystemConfig>,
}
