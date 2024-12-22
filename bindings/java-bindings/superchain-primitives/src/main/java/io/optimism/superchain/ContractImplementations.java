package io.optimism.superchain;

import org.hyperledger.besu.datatypes.Address;

import java.util.HashMap;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;


public class ContractImplementations {

    private Map<String, Address> l1CrossDomainMessenger= new HashMap<>();

    private Map<String, Address> l1ERC721Bridge= new HashMap<>();

    private Map<String, Address> l1StandardBridge= new HashMap<>();

    private Map<String, Address> l2OutputOracle= new HashMap<>();

    private Map<String, Address> optimismMintableERC20Factory= new HashMap<>();

    private Map<String, Address> optimismPortal= new HashMap<>();

    private Map<String, Address> systemConfig= new HashMap<>();
    // Fault Proof Contracts
    private Optional<Map<String, Address>> anchorStateRegistry= Optional.empty();

    private Optional<Map<String, Address>> delayedWETH= Optional.empty();

    private Optional<Map<String, Address>> disputeGameFactory= Optional.empty();

    private Optional<Map<String, Address>> faultDisputeGame= Optional.empty();

    private Optional<Map<String, Address>> mips= Optional.empty();

    private Optional<Map<String, Address>> permissionedDisputeGame= Optional.empty();

    private Optional<Map<String, Address>> preimageOracle= Optional.empty();

    public Map<String, Address> getL1CrossDomainMessenger() {

        return l1CrossDomainMessenger;
    }

    public void setL1CrossDomainMessenger(Map<String, Address> l1CrossDomainMessenger) {
        this.l1CrossDomainMessenger = l1CrossDomainMessenger;
    }

    public Map<String, Address> getL1ERC721Bridge() {
        return l1ERC721Bridge;
    }

    public void setL1ERC721Bridge(Map<String, Address> l1ERC721Bridge) {
        this.l1ERC721Bridge = l1ERC721Bridge;
    }

    public Map<String, Address> getL1StandardBridge() {
        return l1StandardBridge;
    }

    public void setL1StandardBridge(Map<String, Address> l1StandardBridge) {
        this.l1StandardBridge = l1StandardBridge;
    }

    public Map<String, Address> getL2OutputOracle() {
        return l2OutputOracle;
    }

    public void setL2OutputOracle(Map<String, Address> l2OutputOracle) {
        this.l2OutputOracle = l2OutputOracle;
    }

    public Map<String, Address> getOptimismMintableERC20Factory() {
        return optimismMintableERC20Factory;
    }

    public void setOptimismMintableERC20Factory(Map<String, Address> optimismMintableERC20Factory) {
        this.optimismMintableERC20Factory = optimismMintableERC20Factory;
    }

    public Map<String, Address> getOptimismPortal() {
        return optimismPortal;
    }

    public void setOptimismPortal(Map<String, Address> optimismPortal) {
        this.optimismPortal = optimismPortal;
    }

    public Map<String, Address> getSystemConfig() {
        return systemConfig;
    }

    public void setSystemConfig(Map<String, Address> systemConfig) {
        this.systemConfig = systemConfig;
    }

    public Optional<Map<String, Address>> getAnchorStateRegistry() {
        return anchorStateRegistry;
    }

    public void setAnchorStateRegistry(Optional<Map<String, Address>> anchorStateRegistry) {
        this.anchorStateRegistry = anchorStateRegistry;
    }

    public Optional<Map<String, Address>> getDelayedWETH() {
        return delayedWETH;
    }

    public void setDelayedWETH(Optional<Map<String, Address>> delayedWETH) {
        this.delayedWETH = delayedWETH;
    }

    public Optional<Map<String, Address>> getDisputeGameFactory() {
        return disputeGameFactory;
    }

    public void setDisputeGameFactory(Optional<Map<String, Address>> disputeGameFactory) {
        this.disputeGameFactory = disputeGameFactory;
    }

    public Optional<Map<String, Address>> getFaultDisputeGame() {
        return faultDisputeGame;
    }

    public void setFaultDisputeGame(Optional<Map<String, Address>> faultDisputeGame) {
        this.faultDisputeGame = faultDisputeGame;
    }

    public Optional<Map<String, Address>> getMips() {
        return mips;
    }

    public void setMips(Optional<Map<String, Address>> mips) {
        this.mips = mips;
    }

    public Optional<Map<String, Address>> getPermissionedDisputeGame() {
        return permissionedDisputeGame;
    }

    public void setPermissionedDisputeGame(Optional<Map<String, Address>> permissionedDisputeGame) {
        this.permissionedDisputeGame = permissionedDisputeGame;
    }

    public Optional<Map<String, Address>> getPreimageOracle() {
        return preimageOracle;
    }

    public void setPreimageOracle(Optional<Map<String, Address>> preimageOracle) {
        this.preimageOracle = preimageOracle;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof ContractImplementations that)) return false;
        return Objects.equals(getL1CrossDomainMessenger(), that.getL1CrossDomainMessenger()) && Objects.equals(getL1ERC721Bridge(), that.getL1ERC721Bridge()) && Objects.equals(getL1StandardBridge(), that.getL1StandardBridge()) && Objects.equals(getL2OutputOracle(), that.getL2OutputOracle()) && Objects.equals(getOptimismMintableERC20Factory(), that.getOptimismMintableERC20Factory()) && Objects.equals(getOptimismPortal(), that.getOptimismPortal()) && Objects.equals(getSystemConfig(), that.getSystemConfig()) && Objects.equals(getAnchorStateRegistry(), that.getAnchorStateRegistry()) && Objects.equals(getDelayedWETH(), that.getDelayedWETH()) && Objects.equals(getDisputeGameFactory(), that.getDisputeGameFactory()) && Objects.equals(getFaultDisputeGame(), that.getFaultDisputeGame()) && Objects.equals(getMips(), that.getMips()) && Objects.equals(getPermissionedDisputeGame(), that.getPermissionedDisputeGame()) && Objects.equals(getPreimageOracle(), that.getPreimageOracle());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getL1CrossDomainMessenger(), getL1ERC721Bridge(), getL1StandardBridge(), getL2OutputOracle(), getOptimismMintableERC20Factory(), getOptimismPortal(), getSystemConfig(), getAnchorStateRegistry(), getDelayedWETH(), getDisputeGameFactory(), getFaultDisputeGame(), getMips(), getPermissionedDisputeGame(), getPreimageOracle());
    }

    @Override
    public String toString() {
        return "ContractImplementations{" +
                "l1CrossDomainMessenger=" + l1CrossDomainMessenger +
                ", l1ERC721Bridge=" + l1ERC721Bridge +
                ", l1StandardBridge=" + l1StandardBridge +
                ", l2OutputOracle=" + l2OutputOracle +
                ", optimismMintableERC20Factory=" + optimismMintableERC20Factory +
                ", optimismPortal=" + optimismPortal +
                ", systemConfig=" + systemConfig +
                ", anchorStateRegistry=" + anchorStateRegistry +
                ", delayedWETH=" + delayedWETH +
                ", disputeGameFactory=" + disputeGameFactory +
                ", faultDisputeGame=" + faultDisputeGame +
                ", mips=" + mips +
                ", permissionedDisputeGame=" + permissionedDisputeGame +
                ", preimageOracle=" + preimageOracle +
                '}';
    }

    public void merge(ContractImplementations that) {
        this.l1CrossDomainMessenger.putAll(that.getL1CrossDomainMessenger());
        this.l1ERC721Bridge.putAll(that.getL1ERC721Bridge());
        this.l1StandardBridge.putAll(that.getL1StandardBridge());
        this.l2OutputOracle.putAll(that.getL2OutputOracle());
        this.optimismMintableERC20Factory.putAll(that.getOptimismMintableERC20Factory());
        this.optimismPortal.putAll(that.getOptimismPortal());
        this.systemConfig.putAll(that.getSystemConfig());
        this.anchorStateRegistry = Optional.of(this.anchorStateRegistry.orElse(new HashMap<>())).map(map -> {
            map.putAll(that.getAnchorStateRegistry().orElse(new HashMap<>()));
            return map;
        }).filter(map -> !map.isEmpty());
        this.delayedWETH = Optional.of(this.delayedWETH.orElse(new HashMap<>())).map(map -> {
            map.putAll(that.getDelayedWETH().orElse(new HashMap<>()));
            return map;
        }).filter(map -> !map.isEmpty());
        this.disputeGameFactory = Optional.of(this.disputeGameFactory.orElse(new HashMap<>())).map(map -> {
            map.putAll(that.getDisputeGameFactory().orElse(new HashMap<>()));
            return map;
        }).filter(map -> !map.isEmpty());
        this.faultDisputeGame = Optional.of(this.faultDisputeGame.orElse(new HashMap<>())).map(map -> {
            map.putAll(that.getFaultDisputeGame().orElse(new HashMap<>()));
            return map;
        }).filter(map -> !map.isEmpty());
        this.mips = Optional.of(this.mips.orElse(new HashMap<>())).map(map -> {
            map.putAll(that.getMips().orElse(new HashMap<>()));
            return map;
        }).filter(map -> !map.isEmpty());
        this.permissionedDisputeGame = Optional.of(this.permissionedDisputeGame.orElse(new HashMap<>())).map(map -> {
            map.putAll(that.getPermissionedDisputeGame().orElse(new HashMap<>()));
            return map;
        }).filter(map -> !map.isEmpty());
        this.preimageOracle = Optional.of(this.preimageOracle.orElse(new HashMap<>())).map(map -> {
            map.putAll(that.getPreimageOracle().orElse(new HashMap<>()));
            return map;
        }).filter(map -> !map.isEmpty());
    }
}
