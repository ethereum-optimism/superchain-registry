package io.optimism.superchain;

import org.hyperledger.besu.datatypes.Address;

import java.math.BigInteger;
import java.util.Optional;

/**
 * Represents the system configuration of a chain.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public class SystemConfig {

    private Address batcherAddr;

    private String overhead;

    private String scalar;

    private BigInteger gasLimit;

    private Optional<BigInteger> baseFeeScalar = Optional.empty();

    private Optional<BigInteger> blobBaseFeeScalar = Optional.empty();

    /**
     * Gets batcher addr.
     *
     * @return the batcher addr
     */
    public Address getBatcherAddr() {
        return batcherAddr;
    }

    /**
     * Sets batcher addr.
     *
     * @param batcherAddr the batcher addr
     */
    public void setBatcherAddr(Address batcherAddr) {
        this.batcherAddr = batcherAddr;
    }

    /**
     * Gets overhead.
     *
     * @return the overhead
     */
    public String getOverhead() {
        return overhead;
    }

    /**
     * Sets overhead.
     *
     * @param overhead the overhead
     */
    public void setOverhead(String overhead) {
        this.overhead = overhead;
    }

    /**
     * Gets scalar.
     *
     * @return the scalar
     */
    public String getScalar() {
        return scalar;
    }

    /**
     * Sets scalar.
     *
     * @param scalar the scalar
     */
    public void setScalar(String scalar) {
        this.scalar = scalar;
    }

    /**
     * Gets gas limit.
     *
     * @return the gas limit
     */
    public BigInteger getGasLimit() {
        return gasLimit;
    }

    /**
     * Sets gas limit.
     *
     * @param gasLimit the gas limit
     */
    public void setGasLimit(BigInteger gasLimit) {
        this.gasLimit = gasLimit;
    }

    /**
     * Gets base fee scalar.
     *
     * @return the base fee scalar
     */
    public Optional<BigInteger> getBaseFeeScalar() {
        return baseFeeScalar;
    }

    /**
     * Sets base fee scalar.
     *
     * @param baseFeeScalar the base fee scalar
     */
    public void setBaseFeeScalar(Optional<BigInteger> baseFeeScalar) {
        this.baseFeeScalar = baseFeeScalar;
    }

    /**
     * Gets blob base fee scalar.
     *
     * @return the blob base fee scalar
     */
    public Optional<BigInteger> getBlobBaseFeeScalar() {
        return blobBaseFeeScalar;
    }

    /**
     * Sets blob base fee scalar.
     *
     * @param blobBaseFeeScalar the blob base fee scalar
     */
    public void setBlobBaseFeeScalar(Optional<BigInteger> blobBaseFeeScalar) {
        this.blobBaseFeeScalar = blobBaseFeeScalar;
    }

    @Override
    public String toString() {
        return "SystemConfig{" +
                "batcherAddr=" + batcherAddr +
                ", overhead='" + overhead + '\'' +
                ", scalar='" + scalar + '\'' +
                ", gasLimit=" + gasLimit +
                ", baseFeeScalar=" + baseFeeScalar +
                ", blobBaseFeeScalar=" + blobBaseFeeScalar +
                '}';
    }
}
