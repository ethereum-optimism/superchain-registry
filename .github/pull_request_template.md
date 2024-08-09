<!--
This default template is a guide for PRs adding new chains to the registry.
For other types of PRs, please delete this template and write a brief description of your 
code changes and rationale.
 -->

# Adding a New Chain
This PR adds [Chain Name Here] to the registry.

## Checklist
- [ ] I have declared the chain at the appropriate [Superchain Level](../docs/glossary.md#superchain-level-and-rollup-stage).
- [ ] I have run `just validate <chain-id>` locally to ensure all local validation checks pass. 
- [ ] I have run `just codegen` to ensure that the chainlist and other generated files are up-to-date and include my chain.
