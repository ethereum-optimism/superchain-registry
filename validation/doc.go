// Package validation contains logic to confirm that an OP Chain is either a valid Standard Chain
// or a valid Frontier Chain. These checks are provided as Go tests.

// It performs the following checks, for all chains:
// * The Chain ID is unique.
// * There is a valid, public RPC endpoint which returns a valid response.
//   (This is checked implicitly; other checks will fail if we don't have a valid RPC endpoint.)
//
// For Standard Chains, we perform the following additional checks:
//
// * The resource metering config matches the Standard config.
//   This includes the resource limit, elasticity multiplier, base fee maximum change denominator, minimum base fee, maximum base fee, and maximum gas.
// * There is a valid gas price oracle, and the parameters (blob base fee scalar and base feed scalar) are within the permitted range.
// * The L2 output oracle parameters (submission interval, block time and finalization period) all match the values for OP Mainnet.
// * The calculated genesis hash matches the hash declared in the chain’s config file.

package validation
