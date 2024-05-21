# Superchain Registry Contributing Guide
The Superchain Registry repository contains:
* raw ["per-chain" config data](./README.md#3-understand-output) in `yaml` and `json/json.gz` files arranged in a semantically meaningful directory structure
* [superchain-wide config data](#superchain-wide-config-data)
* a Go workspace with
  - a [`superchain`](#superchain-go-module) module
  - a [`validation`](#validation-go-module) module
  - an [`add-chain`](#add-chain-go-module) module
  - The modules are tracked by a top level `go.work` file. The associated `go.work.sum` file is gitignored and not important to typical workflows, which should mirror those of the [CI configuration](.circleci/config.yml).
* a Forge/Solidity script [`CheckSecurityConfigs`](#checksecurityconfigs)
* Automatically generated summary `chainIds.json` file


## Superchain-wide config data
A superchain target defines a set of layer 2 chains which share a `SuperchainConfig` and `ProtocolVersions` contract deployment on layer 1. It is usually named after the layer 1 chain, possibly with an extra identifier to distinguish devnets.


> **Note**
> Example: `sepolia` and `sepolia-dev-0` are distinct superchain targets, although they are on the same layer 1 chain.

### Adding a superchain target
A new Superchain Target can be added by creating a new superchain config directory,
with a `superchain.yaml` config file.

> **Note**
> This is an infrequent operation and unecessary if you are just looking to add a chain to an existing superchain.

Here's an example:

```bash
cd superchain-registry

export SUPERCHAIN_TARGET=goerli-dev-0
mkdir superchain/configs/$SUPERCHAIN_TARGET

cat > superchain/configs/$SUPERCHAIN_TARGET/superchain.yaml << EOF
name: Goerli Dev 0
l1:
  chain_id: 5
  public_rpc: https://ethereum-goerli-rpc.allthatnode.com
  explorer: https://goerli.etherscan.io

protocol_versions_addr: null # todo
superchain_config_addr: null # todo
EOF
```
Superchain-wide configuration, like the `ProtocolVersions` contract address, should be configured here when available.

### Approved contract versions
Each superchain target should have a `semver.yaml` file in the same directory declaring the approved contract semantic versions for that superchain, e.g:
```yaml
l1_cross_domain_messenger: 1.4.0
l1_erc721_bridge: 1.0.0
l1_standard_bridge: 1.1.0
l2_output_oracle: 1.3.0
optimism_mintable_erc20_factory: 1.1.0
optimism_portal: 1.6.0
system_config: 1.3.0

# superchain-wide contracts
protocol_versions: 1.0.0
superchain_config: 1.1.0
```
It is meant to be used when building transactions that upgrade the implementations set in the proxies. See the `semver.yaml` files in existing superchain targets for the latest set of contracts to specify.

### `implementations`
Per superchain a set of canonical implementation deployments, per semver version, is tracked.
As default, an empty collection of deployments can be set:
```bash
cat > superchain/implementations/networks/$SUPERCHAIN_TARGET.yaml << EOF
l1_cross_domain_messenger:
l1_erc721_bridge:
l1_standard_bridge:
l2_output_oracle:
optimism_mintable_erc20_factory:
optimism_portal:
system_config:
EOF
```

## `superchain` Go Module

Per chain and supechain-wide configs and extra data are embedded into the `superchain` go module, which can be imported like so:

```
go get github.com/ethereum-optimism/superchain-registry/superchain@latest
```
The configs are consumed by downstream OP Stack software, i.e. `op-geth` and `op-node`.


## `validation` Go Module
A second module exists in this repo whose purpose is to validate the config exported by the `superchain` module. It is a separate module to avoid import cycles and polluting downstream dependencies with things like `go-ethereum` (which is used in the validation tests).

## `add-chain` Go module
This module contains the CLI tool for generating `superchain` compliant configs and extra data to the registry.

## CheckSecurityConfigs

The `CheckSecurityConfigs.s.sol` script is used in CI to perform
security checks of OP Chains registered in the `superchain`
directory. At high level, it performs checks to ensure privileges are
properly granted to the right addresses. More specifically, it checks
the following privilege grants and role designations:

1. Generic privileges:
   1. Proxy admins. For example, `L1ERC721BridgeProxy` and
      `OptimismMintableERC20FactoryProxy` specify the proxy admin
      addresses who can change their implementations.
   2. Address managers. For example, `ProxyAdmin` specifies the
      address manager it trusts to look up certain addresses by name.
   3. Contract owners. For example, many `Ownable` contracts use this
      role to specify the message senders allowed to make privileged
      calls.
2. Optimism privileged cross-contract calls:
   1. Trusted messengers. For example, `L1ERC721BridgeProxy` and
      `L1StandardBridgeProxy` specify the cross domain messenger
      address they trust with cross domain message sender information.
   2. Trusted bridges. For example,
      `OptimismMintableERC20FactoryProxy` specifies the L1 standard
      bridge it trusts to mint and burn tokens.
   3. Trusted portal. For example, `L1CrossDomainMessengerProxy`
      specifies the portal it trusts to deposit transactions and get
      L2 senders.
   4. Trusted oracles. For example, `OptimismPortalProxy` specifies
      the L2 oracle they trust with the L2 state root information.
      1. After the FPAC upgrade, the `OptimismPortalProxy` specifies the `DisputeGameFactory` they trust rather
      than the legacy `L2OutputOracle` contract.
   5. Trusted system config. For example, `OptimismPortalProxy`
      specifies the system config they trust to get resource config
      from. TODO(issues/37): add checks for the `ResourceMetering`
      contract.
3. Optimism privileged operational roles:
   1. Guardians. This is the role that can pause withdraws in the
      Optimism protocol.
      1. After the FPAC upgrade, the `Guardian` can also blacklist dispute games and change the respected game type
         in the `OptimismPortal`.
   2. Challengers. This is the role that can delete `L2OutputOracleProxy`'s output roots in the Optimism protocol
      1. After the FPAC upgrade, the `CHALLENGER` is a permissionless role in the `FaultDisputeGame`. However,
         in the `PermissionedDisputeGame`, the `CHALLENGER` role is the only party allowed to dispute output proposals
         created by the `PROPOSER` role.

As a result, here is a visualization of all the relationships the
`CheckSecurityConfigs.s.sol` script checks:

``` mermaid
graph TD
  L1ERC721BridgeProxy -- "admin()" --> ProxyAdmin
  L1ERC721BridgeProxy -- "messenger()" --> L1CrossDomainMessengerProxy

  OptimismMintableERC20FactoryProxy -- "admin()" --> ProxyAdmin
  OptimismMintableERC20FactoryProxy -- "BRIDGE()" --> L1StandardBridgeProxy

  ProxyAdmin -- "addressManager()" --> AddressManager
  ProxyAdmin -- "owner()" --> ProxyOwnerMultisig

  L1CrossDomainMessengerProxy -- "PORTAL()" --> OptimismPortalProxy
  L1CrossDomainMessengerProxy -- "addressManager[address(this)]" --> AddressManager

  L1StandardBridgeProxy -- "getOwner()" -->  ProxyAdmin
  L1StandardBridgeProxy -- "messenger()" --> L1CrossDomainMessengerProxy

  AddressManager -- "owner()" -->  ProxyAdmin

  OptimismPortalProxy -- "admin()" --> ProxyAdmin
  OptimismPortalProxy -- "GUARDIAN()" --> GuardianMultisig
  OptimismPortalProxy -- "L2_ORACLE()" --> L2OutputOracleProxy
  OptimismPortalProxy -- "SYSTEM_CONFIG()" --> SystemConfigProxy
  OptimismPortalProxy -- "disputeGameFactory()" --> DisputeGameFactoryProxy

  L2OutputOracleProxy -- "admin()" --> ProxyAdmin
  L2OutputOracleProxy -- "CHALLENGER()" --> ChallengerMultisig

  SystemConfigProxy -- "admin()" --> ProxyAdmin
  SystemConfigProxy -- "owner()" --> SystemConfigOwnerMultisig

  DisputeGameFactoryProxy -- "admin()" --> ProxyAdmin
  DisputeGameFactoryProxy -- "owner()" --> ProxyAdminOwner

  AnchorStateRegistryProxy -- "admin()" --> ProxyAdmin

  DelayedWETHProxy -- "admin()" --> ProxyAdmin
  DelayedWETHProxy -- "owner()" --> ProxyAdminOwner
```

## Setting up your editor for formatting and linting
If you use VSCode, you can place the following in a `settings.json` file in the gitignored `.vscode` directory:

```json
{
    "go.formatTool": "gofumpt",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "gopls": {
        "formatting.gofumpt": true,
    },
}
```


## Links
See [Superchain Upgrades] OP Stack specifications.

[Superchain Upgrades]: https://specs.optimism.io/protocol/superchain-upgrades.html
