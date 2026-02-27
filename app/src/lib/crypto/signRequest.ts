import { ed25519 } from '@noble/curves/ed25519.js';
import { sha256 } from '@noble/hashes/sha2.js';
import { bytesToHex } from '@noble/hashes/utils.js';

export interface SignRequestParams {
    method: string;
    path: string;
    body?: string | null;
    privateKey: Uint8Array;
    publicKey: Uint8Array;
}

export const signRequest = (
    params: SignRequestParams,
): Record<string, string> => {
    const { method, path, body, privateKey, publicKey } = params;

    const timestamp = Math.floor(Date.now() / 1000).toString();
    const nonce = crypto.randomUUID();

    const bodyBytes = new TextEncoder().encode(body || '');
    const bodyHash = bytesToHex(sha256(bodyBytes));

    const payload = `${method}:${path}:${timestamp}:${bodyHash}:${nonce}`;
    const payloadBytes = new TextEncoder().encode(payload);

    const signature = ed25519.sign(payloadBytes, privateKey);

    console.log('[SIGN DEBUG]', {
        method,
        path,
        body: body || '(empty)',
        bodyHash,
        timestamp,
        nonce,
        payload,
        signatureLength: signature.length,
    });

    return {
        'X-Public-Key': bytesToHex(publicKey),
        'X-Signature': bytesToHex(signature),
        'X-Timestamp': timestamp,
        'X-Nonce': nonce,
    };
};
