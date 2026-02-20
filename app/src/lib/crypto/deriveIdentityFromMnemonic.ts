import { mnemonicToSeedSync, validateMnemonic } from '@scure/bip39';
import { wordlist } from '@scure/bip39/wordlists/english.js';
import { ed25519 } from '@noble/curves/ed25519.js';
import { sha256 } from '@noble/hashes/sha2.js';
import { bytesToHex } from '@noble/hashes/utils.js';

/**
 * Represents a derived cryptographic identity from a BIP39 mnemonic phrase.
 */
export interface DerivedIdentity {
    /** Ed25519 private key (32 bytes) - Handle with extreme care! */
    privateKey: Uint8Array;
    /** Ed25519 public key (32 bytes) */
    publicKey: Uint8Array;
    /** Hex-encoded address derived from public key hash (40 characters, 20 bytes) */
    address: string;
}

/**
 * Derives a cryptographic identity (private key, public key, and address) from a BIP39 mnemonic phrase.
 *
 * @param mnemonic - A valid BIP39 mnemonic phrase (typically 12 or 24 words)
 * @returns The derived identity containing privateKey, publicKey, and address
 * @throws {Error} If the mnemonic is invalid, empty, or fails validation
 *
 * @example
 * ```typescript
 * const identity = deriveIdentityFromMnemonic('witch collapse practice feed shame open despair creek road again ice least');
 * console.log(identity.address); // "a1b2c3d4e5f6..."
 * ```
 *
 * @security
 * - The mnemonic should be treated as a master secret
 * - The returned privateKey must be kept strictly in memory and never logged or persisted insecurely
 * - Clear sensitive data from memory when no longer needed
 */
export const deriveIdentityFromMnemonic = (
    mnemonic: string,
): DerivedIdentity => {
    if (!mnemonic || typeof mnemonic !== 'string') {
        throw new Error('Mnemonic must be a non-empty string');
    }

    const normalizedMnemonic = mnemonic.trim();

    if (!normalizedMnemonic) {
        throw new Error('Mnemonic cannot be empty or only whitespace');
    }

    if (!validateMnemonic(normalizedMnemonic, wordlist)) {
        throw new Error(
            'Invalid mnemonic: must be a valid BIP39 phrase with correct word count and checksum',
        );
    }

    try {
        const seed64 = mnemonicToSeedSync(normalizedMnemonic);

        const privateKey = seed64.slice(0, 32);

        const publicKey = ed25519.getPublicKey(privateKey);

        const pubKeyHash = sha256(publicKey);
        const address = bytesToHex(pubKeyHash).slice(0, 40); // 20 bytes = 40 hex characters

        return {
            privateKey, // IMPORTANT: Keep secure; do not expose
            publicKey,
            address,
        };
    } catch (error) {
        throw new Error(
            `Failed to derive identity from mnemonic: ${error instanceof Error ? error.message : String(error)}`,
        );
    }
};
