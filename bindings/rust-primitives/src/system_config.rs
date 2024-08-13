//! System Config Type

use crate::rollup_config::RollupConfig;
use alloy_consensus::{Eip658Value, Receipt};
use alloy_primitives::{address, b256, Address, Log, B256, U256, U64};
use alloy_sol_types::{sol, SolType};
use anyhow::{anyhow, bail, Result};

/// `keccak256("ConfigUpdate(uint256,uint8,bytes)")`
pub const CONFIG_UPDATE_TOPIC: B256 =
    b256!("1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be");

/// The initial version of the system config event log.
pub const CONFIG_UPDATE_EVENT_VERSION_0: B256 = B256::ZERO;

/// System configuration.
#[derive(Debug, Clone, Default, Hash, Eq, PartialEq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
#[cfg_attr(feature = "serde", serde(rename_all = "camelCase"))]
pub struct SystemConfig {
    /// Batcher address
    pub batcher_address: Address,
    /// Fee overhead value
    pub overhead: U256,
    /// Fee scalar value
    pub scalar: U256,
    /// Gas limit value
    pub gas_limit: u64,
    /// Base fee scalar value
    pub base_fee_scalar: Option<u64>,
    /// Blob base fee scalar value
    pub blob_base_fee_scalar: Option<u64>,
}

/// Represents type of update to the system config.
#[derive(Debug, Clone, Copy, Hash, PartialEq, Eq)]
#[repr(u64)]
pub enum SystemConfigUpdateType {
    /// Batcher update type
    Batcher = 0,
    /// Gas config update type
    GasConfig = 1,
    /// Gas limit update type
    GasLimit = 2,
    /// Unsafe block signer update type
    UnsafeBlockSigner = 3,
}

impl TryFrom<u64> for SystemConfigUpdateType {
    type Error = anyhow::Error;

    fn try_from(value: u64) -> core::prelude::v1::Result<Self, Self::Error> {
        match value {
            0 => Ok(SystemConfigUpdateType::Batcher),
            1 => Ok(SystemConfigUpdateType::GasConfig),
            2 => Ok(SystemConfigUpdateType::GasLimit),
            3 => Ok(SystemConfigUpdateType::UnsafeBlockSigner),
            _ => bail!("Invalid SystemConfigUpdateType value: {}", value),
        }
    }
}

impl SystemConfig {
    /// Filters all L1 receipts to find config updates and applies the config updates.
    pub fn update_with_receipts(
        &mut self,
        receipts: &[Receipt],
        rollup_config: &RollupConfig,
        l1_time: u64,
    ) -> Result<()> {
        for receipt in receipts {
            if Eip658Value::Eip658(false) == receipt.status {
                continue;
            }

            receipt.logs.iter().try_for_each(|log| {
                let topics = log.topics();
                if log.address == rollup_config.l1_system_config_address
                    && !topics.is_empty()
                    && topics[0] == CONFIG_UPDATE_TOPIC
                {
                    if let Err(e) = self.process_config_update_log(log, rollup_config, l1_time) {
                        anyhow::bail!("Failed to process config update log: {:?}", e);
                    }
                }
                Ok::<(), anyhow::Error>(())
            })?;
        }
        Ok::<(), anyhow::Error>(())
    }

    /// Decodes an EVM log entry emitted by the system config contract and applies it as a
    /// [SystemConfig] change.
    ///
    /// Parse log data for:
    ///
    /// ```text
    /// event ConfigUpdate(
    ///    uint256 indexed version,
    ///    UpdateType indexed updateType,
    ///    bytes data
    /// );
    /// ```
    fn process_config_update_log(
        &mut self,
        log: &Log,
        rollup_config: &RollupConfig,
        l1_time: u64,
    ) -> Result<()> {
        if log.topics().len() < 3 {
            bail!("Invalid config update log: not enough topics");
        }
        if log.topics()[0] != CONFIG_UPDATE_TOPIC {
            bail!("Invalid config update log: invalid topic");
        }

        // Parse the config update log
        let version = log.topics()[1];
        if version != CONFIG_UPDATE_EVENT_VERSION_0 {
            bail!("Invalid config update log: unsupported version");
        }
        let update_type = u64::from_be_bytes(
            log.topics()[2].as_slice()[24..]
                .try_into()
                .map_err(|_| anyhow!("Failed to convert update type to u64"))?,
        );
        let log_data = log.data.data.as_ref();

        match update_type.try_into()? {
            SystemConfigUpdateType::Batcher => {
                if log_data.len() != 96 {
                    bail!("Invalid config update log: invalid data length");
                }

                let pointer = <sol!(uint64)>::abi_decode(&log_data[0..32], true)
                    .map_err(|_| anyhow!("Failed to decode batcher update log"))?;
                if pointer != 32 {
                    bail!("Invalid config update log: invalid data pointer");
                }
                let length = <sol!(uint64)>::abi_decode(&log_data[32..64], true)
                    .map_err(|_| anyhow!("Failed to decode batcher update log"))?;
                if length != 32 {
                    bail!("Invalid config update log: invalid data length");
                }

                let batcher_address =
                    <sol!(address)>::abi_decode(&log.data.data.as_ref()[64..], true)
                        .map_err(|_| anyhow!("Failed to decode batcher update log"))?;
                self.batcher_address = batcher_address;
            }
            SystemConfigUpdateType::GasConfig => {
                if log_data.len() != 128 {
                    bail!("Invalid config update log: invalid data length");
                }

                let pointer = <sol!(uint64)>::abi_decode(&log_data[0..32], true)
                    .map_err(|_| anyhow!("Invalid config update log: invalid data pointer"))?;
                if pointer != 32 {
                    bail!("Invalid config update log: invalid data pointer");
                }
                let length = <sol!(uint64)>::abi_decode(&log_data[32..64], true)
                    .map_err(|_| anyhow!("Invalid config update log: invalid data length"))?;
                if length != 64 {
                    bail!("Invalid config update log: invalid data length");
                }

                let overhead = <sol!(uint256)>::abi_decode(&log_data[64..96], true)
                    .map_err(|_| anyhow!("Invalid config update log: invalid overhead"))?;
                let scalar = <sol!(uint256)>::abi_decode(&log_data[96..], true)
                    .map_err(|_| anyhow!("Invalid config update log: invalid scalar"))?;

                if rollup_config.is_ecotone_active(l1_time) {
                    if RollupConfig::check_ecotone_l1_system_config_scalar(scalar.to_be_bytes())
                        .is_err()
                    {
                        // ignore invalid scalars, retain the old system-config scalar
                        return Ok(());
                    }

                    // retain the scalar data in encoded form
                    self.scalar = scalar;
                    // zero out the overhead, it will not affect the state-transition after Ecotone
                    self.overhead = U256::ZERO;
                } else {
                    self.scalar = scalar;
                    self.overhead = overhead;
                }
            }
            SystemConfigUpdateType::GasLimit => {
                if log_data.len() != 96 {
                    bail!("Invalid config update log: invalid data length");
                }

                let pointer = <sol!(uint64)>::abi_decode(&log_data[0..32], true)
                    .map_err(|_| anyhow!("Invalid config update log: invalid data pointer"))?;
                if pointer != 32 {
                    bail!("Invalid config update log: invalid data pointer");
                }
                let length = <sol!(uint64)>::abi_decode(&log_data[32..64], true)
                    .map_err(|_| anyhow!("Invalid config update log: invalid data length"))?;
                if length != 32 {
                    bail!("Invalid config update log: invalid data length");
                }

                let gas_limit = <sol!(uint256)>::abi_decode(&log_data[64..], true)
                    .map_err(|_| anyhow!("Invalid config update log: invalid gas limit"))?;
                self.gas_limit = U64::from(gas_limit).saturating_to::<u64>();
            }
            SystemConfigUpdateType::UnsafeBlockSigner => {
                // Ignored in derivation
            }
        }

        Ok(())
    }
}

/// System accounts
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct SystemAccounts {
    /// The address that can deposit attributes
    pub attributes_depositor: Address,
    /// The address of the attributes predeploy
    pub attributes_predeploy: Address,
    /// The address of the fee vault
    pub fee_vault: Address,
}

impl Default for SystemAccounts {
    fn default() -> Self {
        Self {
            attributes_depositor: address!("deaddeaddeaddeaddeaddeaddeaddeaddead0001"),
            attributes_predeploy: address!("4200000000000000000000000000000000000015"),
            fee_vault: address!("4200000000000000000000000000000000000011"),
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use crate::ChainGenesis;
    use alloc::vec;
    use alloy_primitives::{b256, hex, LogData, B256};

    fn mock_rollup_config(system_config: SystemConfig) -> RollupConfig {
        RollupConfig {
            genesis: ChainGenesis {
                system_config: Some(system_config),
                ..Default::default()
            },
            block_time: 2,
            l1_chain_id: 1,
            l2_chain_id: 10,
            regolith_time: Some(0),
            canyon_time: Some(0),
            delta_time: Some(0),
            ecotone_time: Some(10),
            fjord_time: Some(0),
            granite_time: Some(0),
            holocene_time: Some(0),
            blobs_enabled_l1_timestamp: Some(0),
            da_challenge_address: Some(Address::ZERO),
            ..Default::default()
        }
    }

    #[test]
    fn test_system_config_update_batcher_log() {
        const UPDATE_TYPE: B256 =
            b256!("0000000000000000000000000000000000000000000000000000000000000000");

        let mut system_config = SystemConfig::default();
        let rollup_config = mock_rollup_config(system_config.clone());

        let update_log = Log {
            address: Address::ZERO,
            data: LogData::new_unchecked(
                vec![
                    CONFIG_UPDATE_TOPIC,
                    CONFIG_UPDATE_EVENT_VERSION_0,
                    UPDATE_TYPE,
                ],
                hex!("00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000beef").into()
            )
        };

        // Update the batcher address.
        system_config
            .process_config_update_log(&update_log, &rollup_config, 0)
            .unwrap();

        assert_eq!(
            system_config.batcher_address,
            address!("000000000000000000000000000000000000bEEF")
        );
    }

    #[test]
    fn test_system_config_update_gas_config_log() {
        const UPDATE_TYPE: B256 =
            b256!("0000000000000000000000000000000000000000000000000000000000000001");

        let mut system_config = SystemConfig::default();
        let rollup_config = mock_rollup_config(system_config.clone());

        let update_log = Log {
            address: Address::ZERO,
            data: LogData::new_unchecked(
                vec![
                    CONFIG_UPDATE_TOPIC,
                    CONFIG_UPDATE_EVENT_VERSION_0,
                    UPDATE_TYPE,
                ],
                hex!("00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000babe000000000000000000000000000000000000000000000000000000000000beef").into()
            )
        };

        // Update the batcher address.
        system_config
            .process_config_update_log(&update_log, &rollup_config, 0)
            .unwrap();

        assert_eq!(system_config.overhead, U256::from(0xbabe));
        assert_eq!(system_config.scalar, U256::from(0xbeef));
    }

    #[test]
    fn test_system_config_update_gas_config_log_ecotone() {
        const UPDATE_TYPE: B256 =
            b256!("0000000000000000000000000000000000000000000000000000000000000001");

        let mut system_config = SystemConfig::default();
        let rollup_config = mock_rollup_config(system_config.clone());

        let update_log = Log {
            address: Address::ZERO,
            data: LogData::new_unchecked(
                vec![
                    CONFIG_UPDATE_TOPIC,
                    CONFIG_UPDATE_EVENT_VERSION_0,
                    UPDATE_TYPE,
                ],
                hex!("00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000babe000000000000000000000000000000000000000000000000000000000000beef").into()
            )
        };

        // Update the batcher address.
        system_config
            .process_config_update_log(&update_log, &rollup_config, 10)
            .unwrap();

        assert_eq!(system_config.overhead, U256::from(0));
        assert_eq!(system_config.scalar, U256::from(0xbeef));
    }

    #[test]
    fn test_system_config_update_gas_limit_log() {
        const UPDATE_TYPE: B256 =
            b256!("0000000000000000000000000000000000000000000000000000000000000002");

        let mut system_config = SystemConfig::default();
        let rollup_config = mock_rollup_config(system_config.clone());

        let update_log = Log {
            address: Address::ZERO,
            data: LogData::new_unchecked(
                vec![
                    CONFIG_UPDATE_TOPIC,
                    CONFIG_UPDATE_EVENT_VERSION_0,
                    UPDATE_TYPE,
                ],
                hex!("00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000beef").into()
            )
        };

        // Update the batcher address.
        system_config
            .process_config_update_log(&update_log, &rollup_config, 0)
            .unwrap();

        assert_eq!(system_config.gas_limit, 0xbeef_u64);
    }
}
