package io.optimism.superchain;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonUnwrapped;
import org.hyperledger.besu.datatypes.Address;

import java.util.Objects;
import java.util.Optional;

/**
 * Represents the configuration of a chain.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class ChainConfig {

    private String name;

    private long chainId;

    private String publicRpc;

    private String sequencerRpc;

    private String explorer;

    private SuperchainLevel superchainLevel;

    private Optional<Long> superchainTime = Optional.empty();

    private Address batchInboxAddr;

    private ChainGenesis genesis;

    @JsonProperty("superchain")
    @JsonIgnore
    private String superchain;

    @JsonProperty("chain")
    @JsonIgnore
    private String chain;

    @JsonProperty("hardfork_configuration")
    @JsonUnwrapped
    private HardForkConfiguration hardforkConfiguration;

    @JsonProperty("plasma")
    private Optional<PlasmaConfig> plasma = Optional.empty();

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public long getChainId() {
        return chainId;
    }

    public void setChainId(long chainId) {
        this.chainId = chainId;
    }

    public String getPublicRpc() {
        return publicRpc;
    }

    public void setPublicRpc(String publicRpc) {
        this.publicRpc = publicRpc;
    }

    public String getSequencerRpc() {
        return sequencerRpc;
    }

    public void setSequencerRpc(String sequencerRpc) {
        this.sequencerRpc = sequencerRpc;
    }

    public String getExplorer() {
        return explorer;
    }

    public void setExplorer(String explorer) {
        this.explorer = explorer;
    }

    public SuperchainLevel getSuperchainLevel() {
        return superchainLevel;
    }

    public void setSuperchainLevel(SuperchainLevel superchainLevel) {
        this.superchainLevel = superchainLevel;
    }

    public Optional<Long> getSuperchainTime() {
        return superchainTime;
    }

    public void setSuperchainTime(Optional<Long> superchainTime) {
        this.superchainTime = superchainTime;
    }

    public Address getBatchInboxAddr() {
        return batchInboxAddr;
    }

    public void setBatchInboxAddr(Address batchInboxAddr) {
        this.batchInboxAddr = batchInboxAddr;
    }

    public ChainGenesis getGenesis() {
        return genesis;
    }

    public void setGenesis(ChainGenesis genesis) {
        this.genesis = genesis;
    }

    public String getSuperchain() {
        return superchain;
    }

    public void setSuperchain(String superchain) {
        this.superchain = superchain;
    }

    public String getChain() {
        return chain;
    }

    public void setChain(String chain) {
        this.chain = chain;
    }

    public HardForkConfiguration getHardforkConfiguration() {
        return hardforkConfiguration;
    }

    public void setHardforkConfiguration(HardForkConfiguration hardforkConfiguration) {
        this.hardforkConfiguration = hardforkConfiguration;
    }

    public Optional<PlasmaConfig> getPlasma() {
        return plasma;
    }

    public void setPlasma(Optional<PlasmaConfig> plasma) {
        this.plasma = plasma;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof ChainConfig that)) return false;
        return getChainId() == that.getChainId() && Objects.equals(getName(), that.getName()) && Objects.equals(getPublicRpc(), that.getPublicRpc()) && Objects.equals(getSequencerRpc(), that.getSequencerRpc()) && Objects.equals(getExplorer(), that.getExplorer()) && getSuperchainLevel() == that.getSuperchainLevel() && Objects.equals(getSuperchainTime(), that.getSuperchainTime()) && Objects.equals(getBatchInboxAddr(), that.getBatchInboxAddr()) && Objects.equals(getGenesis(), that.getGenesis()) && Objects.equals(getSuperchain(), that.getSuperchain()) && Objects.equals(getChain(), that.getChain()) && Objects.equals(getHardforkConfiguration(), that.getHardforkConfiguration()) && Objects.equals(getPlasma(), that.getPlasma());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getName(), getChainId(), getPublicRpc(), getSequencerRpc(), getExplorer(), getSuperchainLevel(), getSuperchainTime(), getBatchInboxAddr(), getGenesis(), getSuperchain(), getChain(), getHardforkConfiguration(), getPlasma());
    }

    @Override
    public String toString() {
        return "ChainConfig{" +
                "name='" + name + '\'' +
                ", chainId=" + chainId +
                ", publicRpc='" + publicRpc + '\'' +
                ", sequencerRpc='" + sequencerRpc + '\'' +
                ", explorer='" + explorer + '\'' +
                ", superchainLevel=" + superchainLevel +
                ", superchainTime=" + superchainTime +
                ", batchInboxAddr=" + batchInboxAddr +
                ", genesis=" + genesis +
                ", superchain='" + superchain + '\'' +
                ", chain='" + chain + '\'' +
                ", hardforkConfiguration=" + hardforkConfiguration +
                ", plasma=" + plasma +
                '}';
    }

    public void setMissingForkConfigs(HardForkConfiguration defaults) {
        if (this.superchainTime.isEmpty()) {
            return;
        }

        long superchainTime = this.superchainTime.get();

        HardForkConfiguration cfg = this.hardforkConfiguration;

        if (cfg.getCanyonTime().isPresent() && cfg.getCanyonTime().get() > superchainTime) {
            cfg.setCanyonTime(defaults.getCanyonTime());
        }

        if (cfg.getDeltaTime().isPresent() && cfg.getDeltaTime().get() > superchainTime) {
            cfg.setDeltaTime(defaults.getDeltaTime());
        }

        if (cfg.getEcotoneTime().isPresent() && cfg.getEcotoneTime().get() > superchainTime) {
            cfg.setEcotoneTime(defaults.getEcotoneTime());
        }

        if (cfg.getFjordTime().isPresent() && cfg.getFjordTime().get() > superchainTime) {
            cfg.setFjordTime(defaults.getFjordTime());
        }
    }
}
