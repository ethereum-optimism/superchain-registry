package io.optimism.superchain;

import java.util.Objects;
import java.util.Optional;

/**
 * Represents the genesis of a chain.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class ChainGenesis {

    private BlockID l1;

    private BlockID l2;

    private long l2Time;

    private Optional<String> extraData = Optional.empty();

    private Optional<SystemConfig> systemConfig = Optional.empty();

    /**
     * Gets l1.
     *
     * @return the l1
     */
    public BlockID getL1() {
        return l1;
    }

    /**
     * Sets l1.
     *
     * @param l1 the l1
     */
    public void setL1(BlockID l1) {
        this.l1 = l1;
    }

    /**
     * Gets l2.
     *
     * @return the l2
     */
    public BlockID getL2() {
        return l2;
    }

    /**
     * Sets l2.
     *
     * @param l2 the l2
     */
    public void setL2(BlockID l2) {
        this.l2 = l2;
    }

    /**
     * Gets l2 time.
     *
     * @return the l2 time
     */
    public long getL2Time() {
        return l2Time;
    }

    /**
     * Sets l2 time.
     *
     * @param l2Time the l2 time
     */
    public void setL2Time(long l2Time) {
        this.l2Time = l2Time;
    }

    /**
     * Gets extra data.
     *
     * @return the extra data
     */
    public Optional<String> getExtraData() {
        return extraData;
    }

    /**
     * Sets extra data.
     *
     * @param extraData the extra data
     */
    public void setExtraData(Optional<String> extraData) {
        this.extraData = extraData;
    }

    /**
     * Gets system config.
     *
     * @return the system config
     */
    public Optional<SystemConfig> getSystemConfig() {
        return systemConfig;
    }

    /**
     * Sets system config.
     *
     * @param systemConfig the system config
     */
    public void setSystemConfig(Optional<SystemConfig> systemConfig) {
        this.systemConfig = systemConfig;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof ChainGenesis that)) return false;
        return Objects.equals(getL1(), that.getL1()) && Objects.equals(getL2(), that.getL2()) && Objects.equals(getL2Time(), that.getL2Time()) && Objects.equals(getExtraData(), that.getExtraData()) && Objects.equals(getSystemConfig(), that.getSystemConfig());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getL1(), getL2(), getL2Time(), getExtraData(), getSystemConfig());
    }

    @Override
    public String toString() {
        return "ChainGenesis{" +
                "l1=" + l1 +
                ", l2=" + l2 +
                ", l2Time=" + l2Time +
                ", extraData=" + extraData +
                ", systemConfig=" + systemConfig +
                '}';
    }
}
