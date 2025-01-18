# Hardfork activation inheritance behavior

This document specifies how hardfork activations are coordinated between _standard_ chains via the superchain-registry.

Frontier chains can freely choose when to activate hardforks.

If and when a frontier chain becomes a standard chain, a `superchain_time` field is added to that chain's configuration TOML. This field denotes the time after which the chain will receive Superchain upgrades. When that time passes, hardfork activations will be sourced from the chain's `superchain.toml` file. Values are copied from that file into the chain's configuration TOML. This allows a single file to coordinate the hardfork activations of many chains at once.

For a chain to "receive" a particular default (superchain-wide) hardfork activation time, the following conditions must hold:

* It must have the `superchain_time` set. All superchain-wide hardfork activations will be inherited starting from this timestamp. 
* It must not set a non-nil value for this activation time in its individual configuration file
* The default hardfork activation must be set in the superchain-wide configuration file
* The hardfork activation time must be equal to or after the `superchain_time`
* The hardfork activation time must be strictly after the genesis of the chain in question. If it is equal to or before, the chain should receive the "magic value" of `0` for this hardfork.

To "receive" a hardfork activation time for a given chain means that the downstream OP Stack components will apply that activation time when starting up using a network flag specifying the chain in question.

At the time of writing, this is implemented for
* [op-geth](https://docs.optimism.io/builders/node-operators/configuration/base-config#initialization-via-network-flags)
* [op-node](https://docs.optimism.io/builders/node-operators/configuration/base-config#configuring-op-node)

This implies some more conditions which need to hold for a chain to receive the superchain-wide hardfork activation:

* They must be running the above OP Stack software which supports this feature, with the relevant initialization invocations to trigger it
* The software must be up-to-date enough to embed the latest superchain-wide default hardfork activation times as well as the chain's individual configuration file (complete with `superchain_time` field).

> [!CAUTION]
> If (for example) OP Stack components are initialized without the network flags, this will require manual coordination to pass hardfork activation times into the command line invocation of the relevant commands.
