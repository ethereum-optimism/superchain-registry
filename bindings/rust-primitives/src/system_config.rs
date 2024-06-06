//! System Config Type

use alloc::string::String;
use alloy_primitives::Address;

/// System configuration.
#[derive(Debug, Clone, Default)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "camelCase"))]
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
