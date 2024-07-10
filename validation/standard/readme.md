# Standard Rollup Blockspace Charter definition files
This directory contains a number of TOML files which declare various parameters, addresses, contract versions and other data which define a **standard chain** in the sense of the [Standard Rollup Charter](https://gov.optimism.io/t/season-6-draft-standard-rollup-charter/8135).

Where necessary, distinct declarations are made for testnets and devnets to the configuration to vary to a greater extent.

Parameters may be declared to be equal to a pair of values, meaning that the parameter must be within the bounds defined by those values.

The TOML files are embedded into Go bindings, which are in turn references by the validation checks in the parent directory. The entrypoint for those checks is [`validation_test.go`](../validation_test.go).

