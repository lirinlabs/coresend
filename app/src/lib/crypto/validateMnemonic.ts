import { generateMnemonic, validateMnemonic } from '@scure/bip39';
import { wordlist } from '@scure/bip39/wordlists/english.js';

/**
 * Result of seed phrase validation
 */
export interface ValidationResult {
    /** Whether the entire seed phrase is valid */
    valid: boolean;
    /** Error message if validation failed */
    error?: string;
    /** Indices of invalid words */
    invalidIndices?: number[];
}

/**
 * Validation status for a single word
 */
export interface WordValidation {
    /** Index of the word */
    index: number;
    /** The word itself */
    word: string;
    /** Whether the word exists in BIP39 wordlist */
    isValid: boolean;
}

/**
 * Checks if a single word exists in the BIP39 English wordlist.
 *
 * @param word - The word to check
 * @returns True if the word is in the BIP39 wordlist
 *
 * @example
 * ```typescript
 * isValidBip39Word('abandon'); // true
 * isValidBip39Word('xyz123');  // false
 * ```
 */
export const isValidBip39Word = (word: string): boolean => {
    if (!word || typeof word !== 'string') return false;
    const normalized = word.trim().toLowerCase();
    return wordlist.includes(normalized);
};

/**
 * Validates each word individually against the BIP39 wordlist.
 *
 * @param words - Array of seed words to validate
 * @returns Array of validation results for each word
 *
 * @example
 * ```typescript
 * const validations = validateSeedWordsList(['abandon', 'xyz', 'ability']);
 * // [
 * //   { index: 0, word: 'abandon', isValid: true },
 * //   { index: 1, word: 'xyz', isValid: false },
 * //   { index: 2, word: 'ability', isValid: true }
 * // ]
 * ```
 */
export const validateSeedWordsList = (words: string[]): WordValidation[] => {
    return words.map((word, index) => ({
        index,
        word,
        isValid: isValidBip39Word(word),
    }));
};

/**
 * Validates a complete seed phrase.
 * Checks that all words are filled, all words are in the BIP39 wordlist,
 * and the complete phrase passes BIP39 checksum validation.
 *
 * @param words - Array of 12 or 24 seed words
 * @returns Validation result with error message and invalid word indices if applicable
 *
 * @example
 * ```typescript
 * const result = validateSeedWords(['abandon', 'abandon', ...]);
 * if (!result.valid) {
 *   console.error(result.error);
 *   console.log('Invalid words at indices:', result.invalidIndices);
 * }
 * ```
 */
export const validateSeedWords = (words: string[]): ValidationResult => {
    // Check if all words are filled
    const emptyIndices: number[] = [];
    words.forEach((word, index) => {
        if (!word || !word.trim()) {
            emptyIndices.push(index);
        }
    });

    if (emptyIndices.length > 0) {
        return {
            valid: false,
            error: 'Please fill all 12 words',
            invalidIndices: emptyIndices,
        };
    }

    // Check each word against BIP39 wordlist
    const wordValidations = validateSeedWordsList(words);
    const invalidWords = wordValidations.filter((v) => !v.isValid);

    if (invalidWords.length > 0) {
        return {
            valid: false,
            error: 'Invalid seed phrase. Please check your words.',
            invalidIndices: invalidWords.map((v) => v.index),
        };
    }

    // Join words into mnemonic string and validate BIP39 checksum
    const mnemonic = words.join(' ').trim();

    if (!validateMnemonic(mnemonic, wordlist)) {
        // If individual words are valid but checksum fails, return error w/o specific indices
        return {
            valid: false,
            error: 'Invalid seed phrase. Please check your words.',
            invalidIndices: words.map((_, index) => index),
        };
    }

    return { valid: true };
};

/**
 * Generates a new BIP39 mnemonic phrase as an array of words.
 *
 * @param entropyBits - Entropy in bits (128 for 12 words, 256 for 24 words)
 * @returns Array of seed words
 *
 * @example
 * ```typescript
 * const words = generateMnemonicWords(128); // 12 words
 * console.log(words); // ['abandon', 'ability', 'able', ...]
 * ```
 */
export const generateMnemonicWords = (entropyBits: number = 128): string[] => {
    const mnemonic = generateMnemonic(wordlist, entropyBits);
    return mnemonic.split(' ');
};
