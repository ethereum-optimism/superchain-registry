//! Rollup Config Types

use alloy_eips::eip1559::BaseFeeParams;
use alloy_primitives::{address, b256, uint, Address};
use anyhow::{anyhow, Result};

use crate::block::BlockID;
use crate::chain_config::ChainConfig;
use crate::fee_params::{
    base_fee_params, canyon_base_fee_params, BASE_SEPOLIA_BASE_FEE_PARAMS,
    BASE_SEPOLIA_CANYON_BASE_FEE_PARAMS, OP_BASE_FEE_PARAMS, OP_CANYON_BASE_FEE_PARAMS,
    OP_SEPOLIA_BASE_FEE_PARAMS, OP_SEPOLIA_CANYON_BASE_FEE_PARAMS,
};
use crate::genesis::ChainGenesis;
use crate::system_config::SystemConfig;

/// The max rlp bytes per channel for the Bedrock hardfork.
pub const MAX_RLP_BYTES_PER_CHANNEL_BEDROCK: u64 = 10_000_000;

/// The max rlp bytes per channel for the Fjord hardfork.
pub const MAX_RLP_BYTES_PER_CHANNEL_FJORD: u64 = 100_000_000;

/// The max sequencer drift when the Fjord hardfork is active.
pub const FJORD_MAX_SEQUENCER_DRIFT: u64 = 1800;

/// Returns the rollup config for the given chain ID.
pub fn rollup_config_from_chain_id(chain_id: u64) -> Result<RollupConfig> {
    chain_id.try_into()
}

impl TryFrom<u64> for RollupConfig {
    type Error = anyhow::Error;

    fn try_from(chain_id: u64) -> Result<RollupConfig> {
        match chain_id {
            10 => Ok(OP_MAINNET_CONFIG),
            11155420 => Ok(OP_SEPOLIA_CONFIG),
            8453 => Ok(BASE_MAINNET_CONFIG),
            84532 => Ok(BASE_SEPOLIA_CONFIG),
            _ => Err(anyhow!("Unknown chain ID")),
        }
    }
}

/// The Rollup configuration.
#[derive(Debug, Clone, Eq, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct RollupConfig {
    /// The genesis state of the rollup.
    pub genesis: ChainGenesis,
    /// The block time of the L2, in seconds.
    pub block_time: u64,
    /// Sequencer batches may not be more than MaxSequencerDrift seconds after
    /// the L1 timestamp of the sequencing window end.
    ///
    /// Note: When L1 has many 1 second consecutive blocks, and L2 grows at fixed 2 seconds,
    /// the L2 time may still grow beyond this difference.
    ///
    /// Note: After the Fjord hardfork, this value becomes a constant of `1800`.
    pub max_sequencer_drift: u64,
    /// The sequencer window size.
    pub seq_window_size: u64,
    /// Number of L1 blocks between when a channel can be opened and when it can be closed.
    pub channel_timeout: u64,
    /// The L1 chain ID
    pub l1_chain_id: u64,
    /// The L2 chain ID
    pub l2_chain_id: u64,
    /// Base Fee Params
    pub base_fee_params: BaseFeeParams,
    /// Base fee params post-canyon hardfork
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub canyon_base_fee_params: Option<BaseFeeParams>,
    /// `regolith_time` sets the activation time of the Regolith network-upgrade:
    /// a pre-mainnet Bedrock change that addresses findings of the Sherlock contest related to
    /// deposit attributes. "Regolith" is the loose deposited rock that sits on top of Bedrock.
    /// Active if regolith_time != None && L2 block timestamp >= Some(regolith_time), inactive
    /// otherwise.
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub regolith_time: Option<u64>,
    /// `canyon_time` sets the activation time of the Canyon network upgrade.
    /// Active if `canyon_time` != None && L2 block timestamp >= Some(canyon_time), inactive
    /// otherwise.
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub canyon_time: Option<u64>,
    /// `delta_time` sets the activation time of the Delta network upgrade.
    /// Active if `delta_time` != None && L2 block timestamp >= Some(delta_time), inactive
    /// otherwise.
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub delta_time: Option<u64>,
    /// `ecotone_time` sets the activation time of the Ecotone network upgrade.
    /// Active if `ecotone_time` != None && L2 block timestamp >= Some(ecotone_time), inactive
    /// otherwise.
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub ecotone_time: Option<u64>,
    /// `fjord_time` sets the activation time of the Fjord network upgrade.
    /// Active if `fjord_time` != None && L2 block timestamp >= Some(fjord_time), inactive
    /// otherwise.
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub fjord_time: Option<u64>,
    /// `granite_time` sets the activation time for the Granite network upgrade.
    /// Active if `granite_time` != None && L2 block timestamp >= Some(granite_time), inactive
    /// otherwise.
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub granite_time: Option<u64>,
    /// `holocene_time` sets the activation time for the Holocene network upgrade.
    /// Active if `holocene_time` != None && L2 block timestamp >= Some(holocene_time), inactive
    /// otherwise.
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub holocene_time: Option<u64>,
    /// `batch_inbox_address` is the L1 address that batches are sent to.
    pub batch_inbox_address: Address,
    /// `deposit_contract_address` is the L1 address that deposits are sent to.
    pub deposit_contract_address: Address,
    /// `l1_system_config_address` is the L1 address that the system config is stored at.
    pub l1_system_config_address: Address,
    /// `protocol_versions_address` is the L1 address that the protocol versions are stored at.
    pub protocol_versions_address: Address,
    /// The superchain config address.
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub superchain_config_address: Option<Address>,
    /// `blobs_enabled_l1_timestamp` is the timestamp to start reading blobs as a batch data
    /// source. Optional.
    #[cfg_attr(
        feature = "serde",
        serde(rename = "blobs_data", skip_serializing_if = "Option::is_none")
    )]
    pub blobs_enabled_l1_timestamp: Option<u64>,
    /// `da_challenge_address` is the L1 address that the data availability challenge contract is
    /// stored at.
    #[cfg_attr(feature = "serde", serde(skip_serializing_if = "Option::is_none"))]
    pub da_challenge_address: Option<Address>,
}

// Need to manually implement Default because [`BaseFeeParams`] has no Default impl.
impl Default for RollupConfig {
    fn default() -> Self {
        RollupConfig {
            genesis: ChainGenesis::default(),
            block_time: 0,
            max_sequencer_drift: 0,
            seq_window_size: 0,
            channel_timeout: 0,
            l1_chain_id: 0,
            l2_chain_id: 0,
            base_fee_params: OP_BASE_FEE_PARAMS,
            canyon_base_fee_params: None,
            regolith_time: None,
            canyon_time: None,
            delta_time: None,
            ecotone_time: None,
            fjord_time: None,
            granite_time: None,
            holocene_time: None,
            batch_inbox_address: Address::ZERO,
            deposit_contract_address: Address::ZERO,
            l1_system_config_address: Address::ZERO,
            protocol_versions_address: Address::ZERO,
            superchain_config_address: None,
            blobs_enabled_l1_timestamp: None,
            da_challenge_address: None,
        }
    }
}

/// Loads the rollup config for the OP-Stack chain given the chain config and address list.
pub fn load_op_stack_rollup_config(chain_config: &ChainConfig) -> RollupConfig {
    RollupConfig {
        genesis: chain_config.genesis.clone(),
        l1_chain_id: chain_config.l1_chain_id,
        l2_chain_id: chain_config.chain_id,
        base_fee_params: base_fee_params(chain_config.chain_id),
        canyon_base_fee_params: Some(canyon_base_fee_params(chain_config.chain_id)),
        regolith_time: Some(0),
        canyon_time: chain_config.hardfork_configuration.canyon_time,
        delta_time: chain_config.hardfork_configuration.delta_time,
        ecotone_time: chain_config.hardfork_configuration.ecotone_time,
        fjord_time: chain_config.hardfork_configuration.fjord_time,
        granite_time: chain_config.hardfork_configuration.granite_time,
        holocene_time: chain_config.hardfork_configuration.holocene_time,
        batch_inbox_address: chain_config.batch_inbox_addr,
        deposit_contract_address: chain_config
            .addresses
            .as_ref()
            .map(|a| a.optimism_portal_proxy)
            .unwrap_or_default(),
        l1_system_config_address: chain_config
            .addresses
            .as_ref()
            .map(|a| a.system_config_proxy)
            .unwrap_or_default(),
        protocol_versions_address: chain_config
            .addresses
            .as_ref()
            .map(|a| a.address_manager)
            .unwrap_or_default(),
        superchain_config_address: None,
        blobs_enabled_l1_timestamp: None,
        da_challenge_address: chain_config
            .plasma
            .as_ref()
            .and_then(|plasma| plasma.da_challenge_address),

        // The below chain parameters can be different per OP-Stack chain,
        // but since none of the superchain chains differ, it's not represented in the superchain-registry yet.
        // This restriction on superchain-chains may change in the future.
        // Test/Alt configurations can still load custom rollup-configs when necessary.
        block_time: 2,
        channel_timeout: 300,
        max_sequencer_drift: 600,
        seq_window_size: 3600,
    }
}

impl RollupConfig {
    /// Returns true if Regolith is active at the given timestamp.
    pub fn is_regolith_active(&self, timestamp: u64) -> bool {
        self.regolith_time.map_or(false, |t| timestamp >= t)
    }

    /// Returns true if Canyon is active at the given timestamp.
    pub fn is_canyon_active(&self, timestamp: u64) -> bool {
        self.canyon_time.map_or(false, |t| timestamp >= t)
    }

    /// Returns true if Delta is active at the given timestamp.
    pub fn is_delta_active(&self, timestamp: u64) -> bool {
        self.delta_time.map_or(false, |t| timestamp >= t)
    }

    /// Returns true if Ecotone is active at the given timestamp.
    pub fn is_ecotone_active(&self, timestamp: u64) -> bool {
        self.ecotone_time.map_or(false, |t| timestamp >= t)
    }

    /// Returns true if Fjord is active at the given timestamp.
    pub fn is_fjord_active(&self, timestamp: u64) -> bool {
        self.fjord_time.map_or(false, |t| timestamp >= t)
    }

    /// Returns true if Granite is active at the given timestamp.
    pub fn is_granite_active(&self, timestamp: u64) -> bool {
        self.granite_time.map_or(false, |t| timestamp >= t)
    }

    /// Returns true if Holocene is active at the given timestamp.
    pub fn is_holocene_active(&self, timestamp: u64) -> bool {
        self.holocene_time.map_or(false, |t| timestamp >= t)
    }

    /// Returns true if a DA Challenge proxy Address is provided in the rollup config and the
    /// address is not zero.
    pub fn is_plasma_enabled(&self) -> bool {
        self.da_challenge_address
            .map_or(false, |addr| !addr.is_zero())
    }

    /// Returns the max sequencer drift for the given timestamp.
    pub fn max_sequencer_drift(&self, timestamp: u64) -> u64 {
        if self.is_fjord_active(timestamp) {
            FJORD_MAX_SEQUENCER_DRIFT
        } else {
            self.max_sequencer_drift
        }
    }

    /// Returns the max rlp bytes per channel for the given timestamp.
    pub fn max_rlp_bytes_per_channel(&self, timestamp: u64) -> u64 {
        if self.is_fjord_active(timestamp) {
            MAX_RLP_BYTES_PER_CHANNEL_FJORD
        } else {
            MAX_RLP_BYTES_PER_CHANNEL_BEDROCK
        }
    }

    /// Checks the scalar value in Ecotone.
    pub fn check_ecotone_l1_system_config_scalar(scalar: [u8; 32]) -> Result<(), &'static str> {
        let version_byte = scalar[0];
        match version_byte {
            0 => {
                if scalar[1..28] != [0; 27] {
                    return Err("Bedrock scalar padding not empty");
                }
                Ok(())
            }
            1 => {
                if scalar[1..24] != [0; 23] {
                    return Err("Invalid version 1 scalar padding");
                }
                Ok(())
            }
            _ => {
                // ignore the event if it's an unknown scalar format
                Err("Unrecognized scalar version")
            }
        }
    }

    /// Returns the [RollupConfig] for the given L2 chain ID.
    pub fn from_l2_chain_id(l2_chain_id: u64) -> Option<RollupConfig> {
        match l2_chain_id {
            10 => Some(OP_MAINNET_CONFIG),
            11155420 => Some(OP_SEPOLIA_CONFIG),
            8453 => Some(BASE_MAINNET_CONFIG),
            84532 => Some(BASE_SEPOLIA_CONFIG),
            _ => None,
        }
    }
}

/// The [RollupConfig] for OP Mainnet.
pub const OP_MAINNET_CONFIG: RollupConfig = RollupConfig {
    genesis: ChainGenesis {
        l1: BlockID {
            hash: b256!("438335a20d98863a4c0c97999eb2481921ccd28553eac6f913af7c12aec04108"),
            number: 17_422_590_u64,
        },
        l2: BlockID {
            hash: b256!("dbf6a80fef073de06add9b0d14026d6e5a86c85f6d102c36d3d8e9cf89c2afd3"),
            number: 105_235_063_u64,
        },
        l2_time: 1_686_068_903_u64,
        system_config: Some(SystemConfig {
            batcher_address: address!("6887246668a3b87f54deb3b94ba47a6f63f32985"),
            overhead: uint!(0xbc_U256),
            scalar: uint!(0xa6fe0_U256),
            gas_limit: 30_000_000_u64,
            base_fee_scalar: None,
            blob_base_fee_scalar: None,
        }),
        extra_data: None,
    },
    block_time: 2_u64,
    max_sequencer_drift: 600_u64,
    seq_window_size: 3600_u64,
    channel_timeout: 300_u64,
    l1_chain_id: 1_u64,
    l2_chain_id: 10_u64,
    base_fee_params: OP_BASE_FEE_PARAMS,
    canyon_base_fee_params: Some(OP_CANYON_BASE_FEE_PARAMS),
    regolith_time: Some(0_u64),
    canyon_time: Some(1_704_992_401_u64),
    delta_time: Some(1_708_560_000_u64),
    ecotone_time: Some(1_710_374_401_u64),
    fjord_time: Some(1_720_627_201_u64),
    granite_time: None,
    holocene_time: None,
    batch_inbox_address: address!("ff00000000000000000000000000000000000010"),
    deposit_contract_address: address!("beb5fc579115071764c7423a4f12edde41f106ed"),
    l1_system_config_address: address!("229047fed2591dbec1ef1118d64f7af3db9eb290"),
    protocol_versions_address: address!("8062abc286f5e7d9428a0ccb9abd71e50d93b935"),
    superchain_config_address: Some(address!("95703e0982140D16f8ebA6d158FccEde42f04a4C")),
    da_challenge_address: None,
    blobs_enabled_l1_timestamp: None,
};

/// The [RollupConfig] for OP Sepolia.
pub const OP_SEPOLIA_CONFIG: RollupConfig = RollupConfig {
    genesis: ChainGenesis {
        l1: BlockID {
            hash: b256!("48f520cf4ddaf34c8336e6e490632ea3cf1e5e93b0b2bc6e917557e31845371b"),
            number: 4071408,
        },
        l2: BlockID {
            hash: b256!("102de6ffb001480cc9b8b548fd05c34cd4f46ae4aa91759393db90ea0409887d"),
            number: 0,
        },
        l2_time: 1691802540,
        system_config: Some(SystemConfig {
            batcher_address: address!("8f23bb38f531600e5d8fddaaec41f13fab46e98c"),
            overhead: uint!(0xbc_U256),
            scalar: uint!(0xa6fe0_U256),
            gas_limit: 30_000_000,
            base_fee_scalar: None,
            blob_base_fee_scalar: None,
        }),
        extra_data: None,
    },
    block_time: 2,
    max_sequencer_drift: 600,
    seq_window_size: 3600,
    channel_timeout: 300,
    l1_chain_id: 11155111,
    l2_chain_id: 11155420,
    base_fee_params: OP_SEPOLIA_BASE_FEE_PARAMS,
    canyon_base_fee_params: Some(OP_SEPOLIA_CANYON_BASE_FEE_PARAMS),
    regolith_time: Some(0),
    canyon_time: Some(1699981200),
    delta_time: Some(1703203200),
    ecotone_time: Some(1708534800),
    fjord_time: Some(1716998400),
    granite_time: Some(1_723_478_400_u64),
    holocene_time: None,
    batch_inbox_address: address!("ff00000000000000000000000000000011155420"),
    deposit_contract_address: address!("16fc5058f25648194471939df75cf27a2fdc48bc"),
    l1_system_config_address: address!("034edd2a225f7f429a63e0f1d2084b9e0a93b538"),
    protocol_versions_address: address!("79add5713b383daa0a138d3c4780c7a1804a8090"),
    superchain_config_address: Some(address!("C2Be75506d5724086DEB7245bd260Cc9753911Be")),
    da_challenge_address: None,
    blobs_enabled_l1_timestamp: None,
};

/// The [RollupConfig] for Base Mainnet.
pub const BASE_MAINNET_CONFIG: RollupConfig = RollupConfig {
    genesis: ChainGenesis {
        l1: BlockID {
            hash: b256!("5c13d307623a926cd31415036c8b7fa14572f9dac64528e857a470511fc30771"),
            number: 17_481_768_u64,
        },
        l2: BlockID {
            hash: b256!("f712aa9241cc24369b143cf6dce85f0902a9731e70d66818a3a5845b296c73dd"),
            number: 0_u64,
        },
        l2_time: 1686789347_u64,
        system_config: Some(SystemConfig {
            batcher_address: address!("5050f69a9786f081509234f1a7f4684b5e5b76c9"),
            overhead: uint!(0xbc_U256),
            scalar: uint!(0xa6fe0_U256),
            gas_limit: 30_000_000_u64,
            base_fee_scalar: None,
            blob_base_fee_scalar: None,
        }),
        extra_data: None,
    },
    block_time: 2_u64,
    max_sequencer_drift: 600_u64,
    seq_window_size: 3600_u64,
    channel_timeout: 300_u64,
    l1_chain_id: 1_u64,
    l2_chain_id: 8453_u64,
    base_fee_params: OP_BASE_FEE_PARAMS,
    canyon_base_fee_params: Some(OP_CANYON_BASE_FEE_PARAMS),
    regolith_time: Some(0_u64),
    canyon_time: Some(1_704_992_401_u64),
    delta_time: Some(1_708_560_000_u64),
    ecotone_time: Some(1_710_374_401_u64),
    fjord_time: Some(1_720_627_201_u64),
    granite_time: None,
    holocene_time: None,
    batch_inbox_address: address!("ff00000000000000000000000000000000008453"),
    deposit_contract_address: address!("49048044d57e1c92a77f79988d21fa8faf74e97e"),
    l1_system_config_address: address!("73a79fab69143498ed3712e519a88a918e1f4072"),
    protocol_versions_address: address!("8062abc286f5e7d9428a0ccb9abd71e50d93b935"),
    superchain_config_address: Some(address!("95703e0982140D16f8ebA6d158FccEde42f04a4C")),
    da_challenge_address: None,
    blobs_enabled_l1_timestamp: None,
};

/// The [RollupConfig] for Base Sepolia.
pub const BASE_SEPOLIA_CONFIG: RollupConfig = RollupConfig {
    genesis: ChainGenesis {
        l1: BlockID {
            hash: b256!("cac9a83291d4dec146d6f7f69ab2304f23f5be87b1789119a0c5b1e4482444ed"),
            number: 4370868,
        },
        l2: BlockID {
            hash: b256!("0dcc9e089e30b90ddfc55be9a37dd15bc551aeee999d2e2b51414c54eaf934e4"),
            number: 0,
        },
        l2_time: 1695768288,
        system_config: Some(SystemConfig {
            batcher_address: address!("6cdebe940bc0f26850285caca097c11c33103e47"),
            overhead: uint!(0x834_U256),
            scalar: uint!(0xf4240_U256),
            gas_limit: 25000000,
            base_fee_scalar: None,
            blob_base_fee_scalar: None,
        }),
        extra_data: None,
    },
    block_time: 2,
    max_sequencer_drift: 600,
    seq_window_size: 3600,
    channel_timeout: 300,
    l1_chain_id: 11155111,
    l2_chain_id: 84532,
    base_fee_params: BASE_SEPOLIA_BASE_FEE_PARAMS,
    canyon_base_fee_params: Some(BASE_SEPOLIA_CANYON_BASE_FEE_PARAMS),
    regolith_time: Some(0),
    canyon_time: Some(1699981200),
    delta_time: Some(1703203200),
    ecotone_time: Some(1708534800),
    fjord_time: Some(1716998400),
    granite_time: Some(1_723_478_400_u64),
    holocene_time: None,
    batch_inbox_address: address!("ff00000000000000000000000000000000084532"),
    deposit_contract_address: address!("49f53e41452c74589e85ca1677426ba426459e85"),
    l1_system_config_address: address!("f272670eb55e895584501d564afeb048bed26194"),
    protocol_versions_address: address!("79add5713b383daa0a138d3c4780c7a1804a8090"),
    superchain_config_address: Some(address!("C2Be75506d5724086DEB7245bd260Cc9753911Be")),
    da_challenge_address: None,
    blobs_enabled_l1_timestamp: None,
};

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_regolith_active() {
        let mut config = RollupConfig::default();
        assert!(!config.is_regolith_active(0));
        config.regolith_time = Some(10);
        assert!(config.is_regolith_active(10));
        assert!(!config.is_regolith_active(9));
    }

    #[test]
    fn test_canyon_active() {
        let mut config = RollupConfig::default();
        assert!(!config.is_canyon_active(0));
        config.canyon_time = Some(10);
        assert!(config.is_canyon_active(10));
        assert!(!config.is_canyon_active(9));
    }

    #[test]
    fn test_delta_active() {
        let mut config = RollupConfig::default();
        assert!(!config.is_delta_active(0));
        config.delta_time = Some(10);
        assert!(config.is_delta_active(10));
        assert!(!config.is_delta_active(9));
    }

    #[test]
    fn test_ecotone_active() {
        let mut config = RollupConfig::default();
        assert!(!config.is_ecotone_active(0));
        config.ecotone_time = Some(10);
        assert!(config.is_ecotone_active(10));
        assert!(!config.is_ecotone_active(9));
    }

    #[test]
    fn test_fjord_active() {
        let mut config = RollupConfig::default();
        assert!(!config.is_fjord_active(0));
        config.fjord_time = Some(10);
        assert!(config.is_fjord_active(10));
        assert!(!config.is_fjord_active(9));
    }

    #[test]
    fn test_granite_active() {
        let mut config = RollupConfig::default();
        assert!(!config.is_granite_active(0));
        config.granite_time = Some(10);
        assert!(config.is_granite_active(10));
        assert!(!config.is_granite_active(9));
    }

    #[test]
    fn test_holocene_active() {
        let mut config = RollupConfig::default();
        assert!(!config.is_holocene_active(0));
        config.holocene_time = Some(10);
        assert!(config.is_holocene_active(10));
        assert!(!config.is_holocene_active(9));
    }

    #[test]
    fn test_plasma_enabled() {
        let mut config = RollupConfig::default();
        assert!(!config.is_plasma_enabled());
        config.da_challenge_address = Some(Address::ZERO);
        assert!(!config.is_plasma_enabled());
        config.da_challenge_address = Some(address!("0000000000000000000000000000000000000001"));
        assert!(config.is_plasma_enabled());
    }

    #[test]
    fn test_max_sequencer_drift() {
        let mut config = RollupConfig {
            max_sequencer_drift: 100,
            ..Default::default()
        };
        assert_eq!(config.max_sequencer_drift(0), 100);
        config.fjord_time = Some(10);
        assert_eq!(config.max_sequencer_drift(0), 100);
        assert_eq!(config.max_sequencer_drift(10), FJORD_MAX_SEQUENCER_DRIFT);
    }
}
