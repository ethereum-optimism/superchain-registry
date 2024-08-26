#![doc = include_str!("../README.md")]
#![warn(missing_debug_implementations, missing_docs, rustdoc::all)]
#![deny(unused_must_use, rust_2018_idioms)]
#![cfg_attr(docsrs, feature(doc_cfg, doc_auto_cfg))]
#![cfg_attr(not(feature = "std"), no_std)]
#![cfg_attr(not(test), warn(unused_crate_dependencies))]

/// Re-export types from [superchain_primitives].
pub use superchain_primitives::*;

/// Re-export [superchain_registry].
#[cfg(feature = "serde")]
pub use superchain_registry;

#[cfg(feature = "serde")]
pub use superchain_registry::{
    Chain, ChainList, HashMap, Registry, CHAINS, OPCHAINS, ROLLUP_CONFIGS,
};
