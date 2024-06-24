package io.optimism.superchain;

import org.hyperledger.besu.datatypes.Address;

import java.math.BigInteger;
import java.util.Objects;
import java.util.Optional;

/**
 * Represents the Plasma configuration of a chain.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class PlasmaConfig {

    private Optional<Address> daChallengeAddress = Optional.empty();

    private Optional<BigInteger> daChallengeWindow = Optional.empty();

    private Optional<BigInteger> daResolveWindow = Optional.empty();

    /**
     * Gets da challenge address.
     *
     * @return the da challenge address
     */
    public Optional<Address> getDaChallengeAddress() {
        return daChallengeAddress;
    }

    /**
     * Sets da challenge address.
     *
     * @param daChallengeAddress the da challenge address
     */
    public void setDaChallengeAddress(Optional<Address> daChallengeAddress) {
        this.daChallengeAddress = daChallengeAddress;
    }

    /**
     * Gets da challenge window.
     *
     * @return the da challenge window
     */
    public Optional<BigInteger> getDaChallengeWindow() {
        return daChallengeWindow;
    }

    /**
     * Sets da challenge window.
     *
     * @param daChallengeWindow the da challenge window
     */
    public void setDaChallengeWindow(Optional<BigInteger> daChallengeWindow) {
        this.daChallengeWindow = daChallengeWindow;
    }

    /**
     * Gets da resolve window.
     *
     * @return the da resolve window
     */
    public Optional<BigInteger> getDaResolveWindow() {
        return daResolveWindow;
    }

    /**
     * Sets da resolve window.
     *
     * @param daResolveWindow the da resolve window
     */
    public void setDaResolveWindow(Optional<BigInteger> daResolveWindow) {
        this.daResolveWindow = daResolveWindow;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof PlasmaConfig that)) return false;
        return Objects.equals(getDaChallengeAddress(), that.getDaChallengeAddress()) && Objects.equals(getDaChallengeWindow(), that.getDaChallengeWindow()) && Objects.equals(getDaResolveWindow(), that.getDaResolveWindow());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getDaChallengeAddress(), getDaChallengeWindow(), getDaResolveWindow());
    }

    @Override
    public String toString() {
        return "PlasmaConfig{" +
                "daChallengeAddress=" + daChallengeAddress +
                ", daChallengeWindow=" + daChallengeWindow +
                ", daResolveWindow=" + daResolveWindow +
                '}';
    }
}
