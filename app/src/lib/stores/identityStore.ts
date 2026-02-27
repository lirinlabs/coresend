import { create } from 'zustand';
import type { DerivedIdentity } from '../crypto/deriveIdentityFromMnemonic';

interface IdentityState {
    mnemonic: string | null;
    identities: DerivedIdentity[];
    activeIndex: number;
    setMnemonic: (mnemonic: string) => void;
    addIdentity: (identity: DerivedIdentity) => void;
    removeIdentity: (index: number) => void;
    setActiveIndex: (index: number) => void;
    clearAll: () => void;
    getActiveIdentity: () => DerivedIdentity | null;
    getActiveAddress: () => string;
    hasIdentities: () => boolean;
}

export const useIdentityStore = create<IdentityState>((set, get) => ({
    mnemonic: null,
    identities: [],
    activeIndex: 0,

    setMnemonic: (mnemonic) => set({ mnemonic }),

    addIdentity: (identity) =>
        set((state) => {
            if (state.identities.some((i) => i.address === identity.address)) {
                return state;
            }
            return { identities: [...state.identities, identity] };
        }),

    removeIdentity: (index) =>
        set((state) => {
            const newIdentities = state.identities.filter(
                (_, i) => i !== index,
            );
            const newActiveIndex = Math.min(
                state.activeIndex,
                Math.max(0, newIdentities.length - 1),
            );
            return {
                identities: newIdentities,
                activeIndex: newIdentities.length === 0 ? 0 : newActiveIndex,
            };
        }),

    setActiveIndex: (index) =>
        set((state) => {
            if (state.identities.length === 0) return state;
            const clamped = Math.max(
                0,
                Math.min(index, state.identities.length - 1),
            );
            return clamped === state.activeIndex
                ? state
                : { activeIndex: clamped };
        }),

    clearAll: () => set({ mnemonic: null, identities: [], activeIndex: 0 }),

    getActiveIdentity: () => {
        const state = get();
        return state.identities[state.activeIndex] ?? null;
    },

    getActiveAddress: () => {
        const state = get();
        const identity = state.identities[state.activeIndex];
        return identity?.address ?? '';
    },

    hasIdentities: () => {
        const state = get();
        return state.identities.length > 0;
    },
}));
