import { signRequest } from '../crypto/signRequest';
import { useIdentityStore } from '../stores/identityStore';

export const customFetch = async <T>(
    url: string,
    options?: RequestInit,
): Promise<T> => {
    const identity = useIdentityStore.getState().identity;

    if (!identity) {
        throw new Error('No identity - please unlock inbox first');
    }

    let body: string | undefined;
    if (options?.body !== undefined && options?.body !== null) {
        if (typeof options.body === 'string') {
            body = options.body;
        } else if (
            options.body instanceof FormData ||
            options.body instanceof Blob
        ) {
            throw new Error(
                'FormData and Blob body types are not supported for signature authentication',
            );
        } else {
            throw new Error(`Unsupported body type: ${typeof options.body}`);
        }
    }
    const path = new URL(url, window.location.origin).pathname;
    const method = options?.method || 'GET';

    console.log('[AUTH DEBUG]', { url, path, method, body: body ?? '(empty)' });

    const authHeaders: Record<string, string> = signRequest({
        method,
        path,
        body,
        privateKey: identity.privateKey,
        publicKey: identity.publicKey,
    });

    const mergedHeaders = new Headers(options?.headers as HeadersInit);
    for (const [k, v] of Object.entries(authHeaders)) {
        mergedHeaders.set(k, v);
    }

    const response = await fetch(url, {
        ...options,
        headers: mergedHeaders,
    });

    if (!response.ok) {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
            const errorBody = await response.json();
            if (errorBody?.error?.message) {
                errorMessage = `${errorMessage} - ${errorBody.error.message}`;
            }
        } catch {}
        throw new Error(errorMessage);
    }

    const data = await response.json();
    return { data, status: response.status, headers: response.headers } as T;
};
