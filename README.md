# superchain-registry

> ⚠️ This repository is a **work in progress**.  At a later date, it will be proposed to, and must be approved by, Optimism Governance.  Until that time, the configuration described here is subject to change.

The Superchain Registry repository hosts Superchain-configuration data in a minimal human-readable form.
This includes mainnet and testnet Superchain targets, and their respective member chains.

Other configuration, such as contract-permissions and `SystemConfig` parameters are hosted and governed onchain.

The superchain configs are made available in minimal form, to embed in OP-Stack software.
Full deployment artifacts and genesis-states can be derived from the minimal form
using the reference [`op-chain-ops`] tooling.

The `semver.yaml` files each represent the semantic versioning lockfile for the all of the smart contracts in that superchain.
It is meant to be used when building transactions that upgrade the implementations set in the proxies.

If you would like to contribute a new chain or superchain, please see our [contributing guide](./CONTRIBUTING.md).

## Superchain Go Module

Superchain configs can be imported as Go-module:
```
go get github.com/ethereum-optimism/superchain-registry/superchain@latest
```
See [`op-chain-ops`] for config tooling and
 for smart-contract bindings.

[`op-chain-ops`]: https://github.com/ethereum-optimism/optimism/tree/develop/op-chain-ops
[`op-bindings`]: https://github.com/ethereum-optimism/optimism/tree/develop/op-bindings


## Validation Go Module
A second module exists in this repo whose purpose is to validate the config exported by the `superchain` module. It is a separate module to avoid import cycles and polluting downstream dependencies with things like `go-ethereum` (which is used in the validation tests). The modules are tracked by a top level `go.work` file. The associated `go.work.sum` file is gitignored and not important to typical workflows, which should mirror those of the [CI configuration](.circleci/config.yml).

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
   5. Trusted system config. For example, `OptimismPortalProxy`
      specifies the system config they trust to get resource config
      from. TODO(issues/37): add checks for the `ResourceMetering`
      contract.
3. Optimism privileged operational roles:
   1. Guardians. This is the role that can pause withdraws in the
      Optimism protocol.
   2. Challengers. This is the role that can delete
      `L2OutputOracleProxy`'s output roots in the Optimism protocol

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

  L2OutputOracleProxy -- "admin()" --> ProxyAdmin
  L2OutputOracleProxy -- "CHALLENGER()" --> ChallengerMultisig

  SystemConfigProxy -- "admin()" --> ProxyAdmin
  SystemConfigProxy -- "owner()" --> SystemConfigOwnerMultisig

```

## License

MIT License, see [`LICENSE` file](./LICENSE).
