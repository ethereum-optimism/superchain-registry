package io.optimism.superchain;

import java.util.Objects;
import java.util.Optional;

/**
 * Represents the hard fork configuration of a chain.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class HardForkConfiguration {

    private Optional<Long> canyonTime = Optional.empty();

    private Optional<Long> deltaTime = Optional.empty();

    private Optional<Long> ecotoneTime = Optional.empty();

    private Optional<Long> fjordTime = Optional.empty();

    /**
     * Gets canyon time.
     *
     * @return the canyon time
     */
    public Optional<Long> getCanyonTime() {
        return canyonTime;
    }

    /**
     * Sets canyon time.
     *
     * @param canyonTime the canyon time
     */
    public void setCanyonTime(Optional<Long> canyonTime) {
        this.canyonTime = canyonTime;
    }

    /**
     * Gets delta time.
     *
     * @return the delta time
     */
    public Optional<Long> getDeltaTime() {
        return deltaTime;
    }

    /**
     * Sets delta time.
     *
     * @param deltaTime the delta time
     */
    public void setDeltaTime(Optional<Long> deltaTime) {
        this.deltaTime = deltaTime;
    }

    /**
     * Gets ecotone time.
     *
     * @return the ecotone time
     */
    public Optional<Long> getEcotoneTime() {
        return ecotoneTime;
    }

    /**
     * Sets ecotone time.
     *
     * @param ecotoneTime the ecotone time
     */
    public void setEcotoneTime(Optional<Long> ecotoneTime) {
        this.ecotoneTime = ecotoneTime;
    }

    /**
     * Gets fjord time.
     *
     * @return the fjord time
     */
    public Optional<Long> getFjordTime() {
        return fjordTime;
    }

    /**
     * Sets fjord time.
     *
     * @param fjordTime the fjord time
     */
    public void setFjordTime(Optional<Long> fjordTime) {
        this.fjordTime = fjordTime;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof HardForkConfiguration that)) return false;
        return Objects.equals(canyonTime, that.canyonTime) && Objects.equals(deltaTime, that.deltaTime) && Objects.equals(ecotoneTime, that.ecotoneTime) && Objects.equals(fjordTime, that.fjordTime);
    }

    @Override
    public int hashCode() {
        return Objects.hash(canyonTime, deltaTime, ecotoneTime, fjordTime);
    }

    @Override
    public String toString() {
        return "HardForkConfiguration{" +
                "canyonTime=" + canyonTime +
                ", deltaTime=" + deltaTime +
                ", ecotoneTime=" + ecotoneTime +
                ", fjordTime=" + fjordTime +
                '}';
    }
}
