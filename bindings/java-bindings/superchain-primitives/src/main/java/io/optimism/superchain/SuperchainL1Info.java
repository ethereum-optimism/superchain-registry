package io.optimism.superchain;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Objects;

/**
 * The type Superchainl1info.
 *
 * @author grapebaba
 * @since 0.1.0
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public class SuperchainL1Info {

    @JsonProperty("chain_id")
    private long chainId;

    @JsonProperty("public_rpc")
    private String publicRpc;

    @JsonProperty("explorer")
    private String explorer;

    /**
     * Gets chain id.
     *
     * @return the chain id
     */
    public long getChainId() {
        return chainId;
    }

    /**
     * Sets chain id.
     *
     * @param chainId the chain id
     */
    public void setChainId(long chainId) {
        this.chainId = chainId;
    }

    /**
     * Gets public rpc.
     *
     * @return the public rpc
     */
    public String getPublicRpc() {
        return publicRpc;
    }

    /**
     * Sets public rpc.
     *
     * @param publicRpc the public rpc
     */
    public void setPublicRpc(String publicRpc) {
        this.publicRpc = publicRpc;
    }

    /**
     * Gets explorer.
     *
     * @return the explorer
     */
    public String getExplorer() {
        return explorer;
    }

    /**
     * Sets explorer.
     *
     * @param explorer the explorer
     */
    public void setExplorer(String explorer) {
        this.explorer = explorer;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof SuperchainL1Info that)) return false;
        return getChainId() == that.getChainId() && Objects.equals(getPublicRpc(), that.getPublicRpc()) && Objects.equals(getExplorer(), that.getExplorer());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getChainId(), getPublicRpc(), getExplorer());
    }

    @Override
    public String toString() {
        return "SuperchainL1Info{" +
                "chainId=" + chainId +
                ", publicRpc='" + publicRpc + '\'' +
                ", explorer='" + explorer + '\'' +
                '}';
    }
}
