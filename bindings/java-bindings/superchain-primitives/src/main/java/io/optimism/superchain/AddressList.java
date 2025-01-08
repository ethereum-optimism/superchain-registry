package io.optimism.superchain;


import org.hyperledger.besu.datatypes.Address;

import java.util.Objects;
import java.util.Optional;

/**
 * AddressList is a POJO that represents the addresses of the contracts that are deployed on the
 * Superchain. This class is used to deserialize the JSON file that contains the addresses of the
 * contracts.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class AddressList {
    private Address addressManager;

    private Address l1CrossDomainMessengerProxy;

    private Address l1ERC721BridgeProxy;

    private Address l1StandardBridgeProxy;

    private Optional<Address> l2OutputOracleProxy= Optional.empty();

    private Address optimismMintableERC20FactoryProxy;

    private Address optimismPortalProxy;

    private Address systemConfigProxy;

    private Address systemConfigOwner;

    private Address proxyAdmin;

    private Address proxyAdminOwner;

    private Address guardian;

    private Optional<Address> challenger= Optional.empty();

    // Fault Proof Contract Addresses
    private Optional<Address> anchorStateRegistryProxy = Optional.empty();

    private Optional<Address> delayedWETHProxy = Optional.empty();

    private Optional<Address> disputeGameFactoryProxy = Optional.empty();

    private Optional<Address> faultDisputeGame = Optional.empty();

    private Optional<Address> mips = Optional.empty();

    private Optional<Address> permissionedDisputeGame = Optional.empty();

    private Optional<Address> preimageOracle = Optional.empty();

    public Address getAddressManager() {
        return addressManager;
    }

    public void setAddressManager(Address addressManager) {
        this.addressManager = addressManager;
    }

    public Address getL1CrossDomainMessengerProxy() {
        return l1CrossDomainMessengerProxy;
    }

    public void setL1CrossDomainMessengerProxy(Address l1CrossDomainMessengerProxy) {
        this.l1CrossDomainMessengerProxy = l1CrossDomainMessengerProxy;
    }

    public Address getL1ERC721BridgeProxy() {
        return l1ERC721BridgeProxy;
    }

    public void setL1ERC721BridgeProxy(Address l1ERC721BridgeProxy) {
        this.l1ERC721BridgeProxy = l1ERC721BridgeProxy;
    }

    public Address getL1StandardBridgeProxy() {
        return l1StandardBridgeProxy;
    }

    public void setL1StandardBridgeProxy(Address l1StandardBridgeProxy) {
        this.l1StandardBridgeProxy = l1StandardBridgeProxy;
    }

    public Optional<Address> getL2OutputOracleProxy() {
        return l2OutputOracleProxy;
    }

    public void setL2OutputOracleProxy(Optional<Address> l2OutputOracleProxy) {
        this.l2OutputOracleProxy = l2OutputOracleProxy;
    }

    public Address getOptimismMintableERC20FactoryProxy() {
        return optimismMintableERC20FactoryProxy;
    }

    public void setOptimismMintableERC20FactoryProxy(Address optimismMintableERC20FactoryProxy) {
        this.optimismMintableERC20FactoryProxy = optimismMintableERC20FactoryProxy;
    }

    public Address getOptimismPortalProxy() {
        return optimismPortalProxy;
    }

    public void setOptimismPortalProxy(Address optimismPortalProxy) {
        this.optimismPortalProxy = optimismPortalProxy;
    }

    public Address getSystemConfigProxy() {
        return systemConfigProxy;
    }

    public void setSystemConfigProxy(Address systemConfigProxy) {
        this.systemConfigProxy = systemConfigProxy;
    }

    public Address getSystemConfigOwner() {
        return systemConfigOwner;
    }

    public void setSystemConfigOwner(Address systemConfigOwner) {
        this.systemConfigOwner = systemConfigOwner;
    }

    public Address getProxyAdmin() {
        return proxyAdmin;
    }

    public void setProxyAdmin(Address proxyAdmin) {
        this.proxyAdmin = proxyAdmin;
    }

    public Address getProxyAdminOwner() {
        return proxyAdminOwner;
    }

    public void setProxyAdminOwner(Address proxyAdminOwner) {
        this.proxyAdminOwner = proxyAdminOwner;
    }

    public Address getGuardian() {
        return guardian;
    }

    public void setGuardian(Address guardian) {
        this.guardian = guardian;
    }

    public Optional<Address> getChallenger() {
        return challenger;
    }

    public void setChallenger(Optional<Address> challenger) {
        this.challenger = challenger;
    }

    public Optional<Address> getAnchorStateRegistryProxy() {
        return anchorStateRegistryProxy;
    }

    public void setAnchorStateRegistryProxy(Optional<Address> anchorStateRegistryProxy) {
        this.anchorStateRegistryProxy = anchorStateRegistryProxy;
    }

    public Optional<Address> getDelayedWETHProxy() {
        return delayedWETHProxy;
    }

    public void setDelayedWETHProxy(Optional<Address> delayedWETHProxy) {
        this.delayedWETHProxy = delayedWETHProxy;
    }

    public Optional<Address> getDisputeGameFactoryProxy() {
        return disputeGameFactoryProxy;
    }

    public void setDisputeGameFactoryProxy(Optional<Address> disputeGameFactoryProxy) {
        this.disputeGameFactoryProxy = disputeGameFactoryProxy;
    }

    public Optional<Address> getFaultDisputeGame() {
        return faultDisputeGame;
    }

    public void setFaultDisputeGame(Optional<Address> faultDisputeGame) {
        this.faultDisputeGame = faultDisputeGame;
    }

    public Optional<Address> getMips() {
        return mips;
    }

    public void setMips(Optional<Address> mips) {
        this.mips = mips;
    }

    public Optional<Address> getPermissionedDisputeGame() {
        return permissionedDisputeGame;
    }

    public void setPermissionedDisputeGame(Optional<Address> permissionedDisputeGame) {
        this.permissionedDisputeGame = permissionedDisputeGame;
    }

    public Optional<Address> getPreimageOracle() {
        return preimageOracle;
    }

    public void setPreimageOracle(Optional<Address> preimageOracle) {
        this.preimageOracle = preimageOracle;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof AddressList that)) return false;
        return Objects.equals(getAddressManager(), that.getAddressManager()) && Objects.equals(getL1CrossDomainMessengerProxy(), that.getL1CrossDomainMessengerProxy()) && Objects.equals(getL1ERC721BridgeProxy(), that.getL1ERC721BridgeProxy()) && Objects.equals(getL1StandardBridgeProxy(), that.getL1StandardBridgeProxy()) && Objects.equals(getL2OutputOracleProxy(), that.getL2OutputOracleProxy()) && Objects.equals(getOptimismMintableERC20FactoryProxy(), that.getOptimismMintableERC20FactoryProxy()) && Objects.equals(getOptimismPortalProxy(), that.getOptimismPortalProxy()) && Objects.equals(getSystemConfigProxy(), that.getSystemConfigProxy()) && Objects.equals(getSystemConfigOwner(), that.getSystemConfigOwner()) && Objects.equals(getProxyAdmin(), that.getProxyAdmin()) && Objects.equals(getProxyAdminOwner(), that.getProxyAdminOwner()) && Objects.equals(getGuardian(), that.getGuardian()) && Objects.equals(getChallenger(), that.getChallenger()) && Objects.equals(getAnchorStateRegistryProxy(), that.getAnchorStateRegistryProxy()) && Objects.equals(getDelayedWETHProxy(), that.getDelayedWETHProxy()) && Objects.equals(getDisputeGameFactoryProxy(), that.getDisputeGameFactoryProxy()) && Objects.equals(getFaultDisputeGame(), that.getFaultDisputeGame()) && Objects.equals(getMips(), that.getMips()) && Objects.equals(getPermissionedDisputeGame(), that.getPermissionedDisputeGame()) && Objects.equals(getPreimageOracle(), that.getPreimageOracle());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getAddressManager(), getL1CrossDomainMessengerProxy(), getL1ERC721BridgeProxy(), getL1StandardBridgeProxy(), getL2OutputOracleProxy(), getOptimismMintableERC20FactoryProxy(), getOptimismPortalProxy(), getSystemConfigProxy(), getSystemConfigOwner(), getProxyAdmin(), getProxyAdminOwner(), getGuardian(), getChallenger(), getAnchorStateRegistryProxy(), getDelayedWETHProxy(), getDisputeGameFactoryProxy(), getFaultDisputeGame(), getMips(), getPermissionedDisputeGame(), getPreimageOracle());
    }

    @Override
    public String toString() {
        return "AddressList{" +
                "addressManager=" + addressManager +
                ", l1CrossDomainMessengerProxy=" + l1CrossDomainMessengerProxy +
                ", l1ERC721BridgeProxy=" + l1ERC721BridgeProxy +
                ", l1StandardBridgeProxy=" + l1StandardBridgeProxy +
                ", l2OutputOracleProxy=" + l2OutputOracleProxy +
                ", optimismMintableERC20FactoryProxy=" + optimismMintableERC20FactoryProxy +
                ", optimismPortalProxy=" + optimismPortalProxy +
                ", systemConfigProxy=" + systemConfigProxy +
                ", systemConfigOwner=" + systemConfigOwner +
                ", proxyAdmin=" + proxyAdmin +
                ", proxyAdminOwner=" + proxyAdminOwner +
                ", guardian=" + guardian +
                ", challenger=" + challenger +
                ", anchorStateRegistryProxy=" + anchorStateRegistryProxy +
                ", delayedWETHProxy=" + delayedWETHProxy +
                ", disputeGameFactoryProxy=" + disputeGameFactoryProxy +
                ", faultDisputeGame=" + faultDisputeGame +
                ", mips=" + mips +
                ", permissionedDisputeGame=" + permissionedDisputeGame +
                ", preimageOracle=" + preimageOracle +
                '}';
    }
}
