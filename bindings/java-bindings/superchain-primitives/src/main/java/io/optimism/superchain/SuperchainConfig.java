package io.optimism.superchain;

import org.hyperledger.besu.datatypes.Address;

import java.util.Objects;
import java.util.Optional;

/**
 * The type Superchainl1info.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class SuperchainConfig {

    private String name;

    private SuperchainL1Info l1;

    private Optional<Address> protocolVersionsAddr;

    private Optional<Address> superchainConfigAddr;

    private HardForkConfiguration hardforkDefaults;

    /**
     * Gets name.
     *
     * @return the name
     */
    public String getName() {
        return name;
    }

    /**
     * Sets name.
     *
     * @param name the name
     */
    public void setName(String name) {
        this.name = name;
    }

    /**
     * Gets l 1.
     *
     * @return the l 1
     */
    public SuperchainL1Info getL1() {
        return l1;
    }

    /**
     * Sets l 1.
     *
     * @param l1 the l 1
     */
    public void setL1(SuperchainL1Info l1) {
        this.l1 = l1;
    }

    /**
     * Gets protocol versions addr.
     *
     * @return the protocol versions addr
     */
    public Optional<Address> getProtocolVersionsAddr() {
        return protocolVersionsAddr;
    }

    /**
     * Sets protocol versions addr.
     *
     * @param protocolVersionsAddr the protocol versions addr
     */
    public void setProtocolVersionsAddr(Optional<Address> protocolVersionsAddr) {
        this.protocolVersionsAddr = protocolVersionsAddr;
    }

    /**
     * Gets superchain config addr.
     *
     * @return the superchain config addr
     */
    public Optional<Address> getSuperchainConfigAddr() {
        return superchainConfigAddr;
    }

    /**
     * Sets superchain config addr.
     *
     * @param superchainConfigAddr the superchain config addr
     */
    public void setSuperchainConfigAddr(Optional<Address> superchainConfigAddr) {
        this.superchainConfigAddr = superchainConfigAddr;
    }

    /**
     * Gets hardfork defaults.
     *
     * @return the hardfork defaults
     */
    public HardForkConfiguration getHardforkDefaults() {
        return hardforkDefaults;
    }

    /**
     * Sets hardfork defaults.
     *
     * @param hardforkDefaults the hardfork defaults
     */
    public void setHardforkDefaults(HardForkConfiguration hardforkDefaults) {
        this.hardforkDefaults = hardforkDefaults;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof SuperchainConfig that)) return false;
        return Objects.equals(getName(), that.getName()) && Objects.equals(getL1(), that.getL1()) && Objects.equals(getProtocolVersionsAddr(), that.getProtocolVersionsAddr()) && Objects.equals(getSuperchainConfigAddr(), that.getSuperchainConfigAddr()) && Objects.equals(getHardforkDefaults(), that.getHardforkDefaults());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getName(), getL1(), getProtocolVersionsAddr(), getSuperchainConfigAddr(), getHardforkDefaults());
    }

    @Override
    public String toString() {
        return "SuperchainConfig{" +
                "name='" + name + '\'' +
                ", l1=" + l1 +
                ", protocolVersionsAddr=" + protocolVersionsAddr +
                ", superchainConfigAddr=" + superchainConfigAddr +
                ", hardforkDefaults=" + hardforkDefaults +
                '}';
    }
}
