import { mnemonicToSeedSync, validateMnemonic } from '@scure/bip39';
import { wordlist } from '@scure/bip39/wordlists/english.js';
import { HDKey } from '@scure/bip32';
import { ed25519 } from '@noble/curves/ed25519.js';
import { sha256 } from '@noble/hashes/sha2.js';
import { bytesToHex } from '@noble/hashes/utils.js';

export interface DerivedIdentity {
    privateKey: Uint8Array;
    publicKey: Uint8Array;
    address: string;
    index: number;
}

export const deriveIdentityFromMnemonic = (
    mnemonic: string,
    index: number = 0,
): DerivedIdentity => {
    if (!mnemonic || typeof mnemonic !== 'string') {
        throw new Error('Mnemonic must be a non-empty string');
    }

    if (typeof index !== 'number' || index < 0 || !Number.isInteger(index)) {
        throw new Error('Index must be a non-negative integer');
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
        const seed = mnemonicToSeedSync(normalizedMnemonic);
        const hdKey = HDKey.fromMasterSeed(seed);
        const childKey = hdKey.derive(`m/44'/0'/${index}'/0/0`);

        if (!childKey.privateKey) {
            throw new Error('Failed to derive private key from HD node');
        }

        const privateKey = childKey.privateKey;
        const publicKey = ed25519.getPublicKey(privateKey);
        const pubKeyHash = sha256(publicKey);
        const address = bytesToHex(pubKeyHash).slice(0, 40);

        return {
            privateKey,
            publicKey,
            address,
            index,
        };
    } catch (error) {
        throw new Error(
            `Failed to derive identity from mnemonic: ${error instanceof Error ? error.message : String(error)}`,
        );
    }
};
