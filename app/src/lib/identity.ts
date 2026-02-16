import { generateMnemonic as generateBip39Mnemonic } from 'bip39';

const WORD_COUNT_12_ENTROPY_BITS = 128;

export function generateMnemonic(): string {
    return generateBip39Mnemonic(WORD_COUNT_12_ENTROPY_BITS);
}

export function mnemonicToWords(mnemonic: string): string[] {
    return mnemonic.trim().toLowerCase().split(/\s+/).filter(Boolean);
}
