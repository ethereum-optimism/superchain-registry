use include_dir::{include_dir, Dir};

/// Directory containing the configuration files for the superchain.
pub(crate) static CONFIGS_DIR: Dir<'_> = include_dir!("$SUPERCHAIN_CONFIGS");

/// Directory containing the extra files for the superchain.
pub(crate) static EXTRA_DIR: Dir<'_> = include_dir!("$SUPERCHAIN_EXTRA");

/// Directory containing the implementation addresses for the superchain.
pub(crate) static IMPLEMENTATIONS_DIR: Dir<'_> = include_dir!("$SUPERCHAIN_IMPLEMENTATIONS");
