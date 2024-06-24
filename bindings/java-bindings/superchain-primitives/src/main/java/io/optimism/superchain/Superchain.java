package io.optimism.superchain;

import java.util.List;
import java.util.Objects;

/**
 * The type Superchain.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class Superchain {

    private SuperchainConfig config;

    private List<Long> chainIds;

    private String superchain;

    public Superchain() {
    }

    public Superchain(SuperchainConfig config, List<Long> chainIds, String superchain) {
        this.config = config;
        this.chainIds = chainIds;
        this.superchain = superchain;
    }

    public SuperchainConfig getConfig() {
        return config;
    }

    public void setConfig(SuperchainConfig config) {
        this.config = config;
    }

    public List<Long> getChainIds() {
        return chainIds;
    }

    public void setChainIds(List<Long> chainIds) {
        this.chainIds = chainIds;
    }

    public String getSuperchain() {
        return superchain;
    }

    public void setSuperchain(String superchain) {
        this.superchain = superchain;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof Superchain that)) return false;
        return Objects.equals(getConfig(), that.getConfig()) && Objects.equals(getChainIds(), that.getChainIds()) && Objects.equals(getSuperchain(), that.getSuperchain());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getConfig(), getChainIds(), getSuperchain());
    }

    @Override
    public String toString() {
        return "Superchain{" +
                "config=" + config +
                ", chainIds=" + chainIds +
                ", superchain='" + superchain + '\'' +
                '}';
    }
}
