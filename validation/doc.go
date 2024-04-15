// Package validation contains logic to confirm that an OP Chain is either a valid Standard Chain
// or a valid Frontier Chain. These checks are provided as Go tests.

// It performs the following checks:

// For Standard Chains:

// * The Chain ID is unique.
// * There is a valid, public RPC endpoint which returns a valid response.
// * The resource metering config matches the Standard config.
//   This includes the resource limit, elasticity multiplier, base fee maximum change denominator, minimum base fee, maximum base fee, and maximum gas.
// * There is a valid gas price oracle, and the parameters (blob base fee scalar and base feed scalar) are within the permitted range.
// * The L2 output oracle parameters (submission interval, block time and finalization period) all match the Standard config.
// * The calculated genesis hash matches the hash declared in the chain’s config file.

// For Frontier Chains:
// * The Chain ID is unique.
// * There is a valid, public RPC endpoint which returns a valid response.

package validation
