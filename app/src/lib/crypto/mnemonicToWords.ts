export function mnemonicToWords(mnemonic: string): string[] {
    return mnemonic.trim().toLowerCase().split(/\s+/).filter(Boolean);
}
