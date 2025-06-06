<!--
This default template is a guide for PRs adding new chains to the registry.
For other types of PRs, please delete this template and write a brief description of your
code changes and rationale.
 -->

# Adding a New Chain

This PR adds [Chain Name Here] to the registry.

## Checklist

- [ ] I have declared the chain at the appropriate [Superchain Level](../docs/glossary.md#superchain-level-and-rollup-stage).
- [ ] I have run `just codegen $SEPOLIA_RPC_URL,$MAINNET_RPC_URL` to ensure that all auto-generated files are updated
- [ ] I have checked `Allow edits from maintainers` on this PR.
