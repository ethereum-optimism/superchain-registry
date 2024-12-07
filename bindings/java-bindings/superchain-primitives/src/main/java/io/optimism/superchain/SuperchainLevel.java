package io.optimism.superchain;

/**
 * Represents the level of a superchain.
 *
 * @author grapebaba
 * @since 0.1.0
 */
public enum SuperchainLevel {
    /**
     * Frontier superchain level.
     */
    Frontier((byte) 1),
    /**
     * Standard superchain level.
     */
    Standard((byte) 2);

    private final byte value;

    SuperchainLevel(byte value) {
        this.value = value;
    }

    /**
     * Gets value.
     *
     * @return the value
     */
    public byte getValue() {
        return value;
    }

    public static SuperchainLevel fromValue(byte value) {
        for (SuperchainLevel level : SuperchainLevel.values()) {
            if (level.getValue() == value) {
                return level;
            }
        }
        return null;
    }

}
