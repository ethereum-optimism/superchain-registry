package io.optimism.superchain;


import org.hyperledger.besu.datatypes.Hash;

import java.math.BigInteger;
import java.util.Objects;

public class BlockID {

    private Hash hash;

    private BigInteger number;

    public Hash getHash() {
        return hash;
    }

    public void setHash(Hash hash) {
        this.hash = hash;
    }

    public BigInteger getNumber() {
        return number;
    }

    public void setNumber(BigInteger number) {
        this.number = number;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof BlockID blockID)) return false;
        return Objects.equals(getHash(), blockID.getHash()) && Objects.equals(getNumber(), blockID.getNumber());
    }

    @Override
    public int hashCode() {
        return Objects.hash(getHash(), getNumber());
    }

    @Override
    public String toString() {
        return "BlockID{" +
                "hash=" + hash +
                ", number=" + number +
                '}';
    }
}
