use std::env;
use std::path::PathBuf;

fn main() {
    let manifest_dir = env::var("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR is not set");

    let superchain_configs = PathBuf::from(&manifest_dir).join("../../superchain/configs");
    let superchain_extra = PathBuf::from(&manifest_dir).join("../../superchain/extra");
    let superchain_implementations = PathBuf::from(&manifest_dir).join("../../superchain/implementations");

    println!("cargo:rerun-if-env-changed=CARGO_MANIFEST_DIR");
    println!("cargo:rerun-if-changed={}", superchain_configs.display());
    println!("cargo:rerun-if-changed={}", superchain_extra.display());
    println!("cargo:rerun-if-changed={}", superchain_implementations.display());

    println!("cargo:rustc-env=SUPERCHAIN_CONFIGS={}", superchain_configs.display());
    println!("cargo:rustc-env=SUPERCHAIN_EXTRA={}", superchain_extra.display());
    println!("cargo:rustc-env=SUPERCHAIN_IMPLEMENTATIONS={}", superchain_implementations.display());
}

