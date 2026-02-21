import { create } from 'zustand';
import type { DerivedIdentity } from '../crypto/deriveIdentityFromMnemonic';

interface IdentityState {
    identity: DerivedIdentity | null;
    setIdentity: (identity: DerivedIdentity) => void;
    clearIdentity: () => void;
}

export const useIdentityStore = create<IdentityState>((set) => ({
    identity: null,
    setIdentity: (identity) => set({ identity }),
    clearIdentity: () => set({ identity: null }),
}));
