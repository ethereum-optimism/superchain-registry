# configs

The main configuration source, with genesis data, and address of onchain system configuration.

This is a collection of extra configuration data because not all address and system-config information is available through L1 for some chains.

Each chain configuration file in this directory contains configuration data that was generated from the [add-chain](../../add-chain/) code.

- name
- chain_id 
- public_rpc
- sequencer_rpc
- explorer
- superchain_level
- standard_chain_candidate
- superchain_time
- system_config_addr
- batch_inbox_addr
- genesis:
  - l1
    - hash
    - number
  - l2:
    - hash
    - number
  - l2_time
- block_time
- seq_window_size

