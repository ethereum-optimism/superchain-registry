//! Block Types

use alloy_primitives::B256;
use core::fmt::Display;

/// Block identifier.
#[derive(Debug, Clone, Copy, Eq, Hash, PartialEq, Default)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct BlockID {
    /// Block hash
    pub hash: B256,
    /// Block number
    pub number: u64,
}

impl Display for BlockID {
    fn fmt(&self, f: &mut core::fmt::Formatter<'_>) -> core::fmt::Result {
        write!(
            f,
            "BlockID {{ hash: {}, number: {} }}",
            self.hash, self.number
        )
    }
}
