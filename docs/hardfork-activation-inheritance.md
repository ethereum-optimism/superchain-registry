# Hardfork activation inheritance behavior

This documents specifies how hardfork activations are coordinated between _standard_ chains via the superchain-registry.

Frontier chains can freely choose when to activate hardforks.

If and when a frontier chain becomes a standard chain, a `superchain_time` key/value pair is added to the configuration file of that chain in the superchain-registry. It denotes the "time when the chain became a part of the standard superchain".

There is a mechanism whereby hardfork activation times are set _superchain-wide_ in the `superchain.toml` configuration file present for each superchain target, but then conditionally propagated to all standard chains in that superchain when configurations are loaded by OP Stack software. This allows a single location to be updated and to coordinate many chains' hardfork activations.

For a chain to "receive" a particular default hardfork activation time, the following conditions must hold:
* It must be a standard chain with `superchain_time` set (only standard chains have this field)
* It must not set a non-nil value for this activation time in its individual configuration file
* The default hardfork activation must be set in the superchain-wide configuration file
* The hardfork activation time must be equal to or after the `superchain_time`
* The hardfork activation time must be strictly after the genesis of the chain in question. If it is equal to or before, the chain should receive the "magic value" of `0` for this hardfork.

To "receive" a hardfork activation time for a given chain means that the downstream OP Stack components will apply that activation time when starting up using a network flag specifying the chain in question.

At the time of writing, this is implemented for
* [op-geth](https://docs.optimism.io/builders/node-operators/configuration/base-config#initialization-via-network-flags)
* [op-node](https://docs.optimism.io/builders/node-operators/configuration/base-config#configuring-op-node)

These components load configuration via the Go bindings in the `superchain` module. See this [init-time code](../superchain/superchain.go#L163-L205) and [tests](../superchain/superchain_test.go#L226-L308).
