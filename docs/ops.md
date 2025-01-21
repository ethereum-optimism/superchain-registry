# Superchain Registry Operations

## Adding a Chain

The following are the steps you need to take to add a chain to the registry:

> [!IMPORTANT]
> Ensure your chain is listed at [ethereum-lists/chains](https://github.com/ethereum-lists/chains).
> This is to ensure your chain has a unique chain ID. Our validation suite will
> check against this repository. This is a mandatory prerequisite before being
> added to the registry.

### 0. Deploy your chain with `op-deployer`

Adding a chain to the Superchain Registry requires using `op-deployer` to deploy your chain. Check out
the [docs](https://docs.optimism.io/builders/chain-operators/tools/op-deployer) for more information on how to use it.

### 1. Install dependencies

Install the following dependencies:

| Dependency                                                            | Version   | Version Check Command |
|-----------------------------------------------------------------------|-----------|-----------------------|
| [git](https://git-scm.com/)                                           | `^2`      | `git --version`       |
| [go](https://go.dev/)                                                 | `^1.21`   | `go version`          |
| [just](https://github.com/casey/just?tab=readme-ov-file#installation) | `^1.28.0` | `just --version`      |

### 2. Fork this repository

You will be raising a Pull Request from your fork to the upstream repo.

We recommend only adding one chain at a time, and starting with a fresh branch of this repo for every chain.

### 3. Generate a config from your state file

From the root of this directory, run the following:

```
just inflate-config <shortname> <path to your state file>
```

This will put two files in the `.staging` directory - `<shortname>.toml` and `<shortname>.json.zst`

### 4. Update generated config

Update `<shortname>.toml` with your chain's information. Most of the data is automatically populated from your deployer
state file, however you will need to populate the following yourself:

- `name`: Human-readable name for your chain.
- `superchain`: Which superchain your chain belongs to. Can be `sepolia` or `mainnet`.
- `public_rpc`: Public RPC endpoint for your chain.
- `sequencer_rpc`: Sequencer RPC endpoint for your chain.
- `explorer`: Block explorer for your chain.
- `deployment_tx_hash`: Transaction hash of the transaction OP Deployer generated to deploy your chain. You'll need
  to look this up on Etherscan for now. It will be automatically populated in the future.

Don't forget to double-check the config for any inaccuracies.

### 5. Commit

Commit your changes to your fork, then open a pull request. When opening your PR:

- Open it from a non-protected branch in your fork (e.g. avoid the `main` branch). This allows maintainers to push to
  your branch if needed, which streamlines the review and merge process.
- Open one PR per chain you would like to add. This ensures the merge of one chain is not blocked by unexpected issues.
- Once the PR is opened, please check the box
  to [allow edits from maintainers](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/allowing-changes-to-a-pull-request-branch-created-from-a-fork).

Automated checks will run on your PR to validate your chain. A report will also be generated that describes your chain's
level of compliance with the standard blockchain charter.

### 6. Await Review

A member of our team will review your PR. When ready, we will generate code from your PR and push to your fork.

> [!IMPORTANT]
> Don't run codegen yourself. This will slow down review.
