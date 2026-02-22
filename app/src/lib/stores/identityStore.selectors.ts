import { useShallow } from 'zustand/react/shallow';
import { useIdentityStore } from './identityStore';

export const useInboxHeaderStore = () => {
    return useIdentityStore(
        useShallow((state) => ({
            identities: state.identities,
            activeIndex: state.activeIndex,
            activeIdentity: state.identities[state.activeIndex] ?? null,
            removeIdentity: state.removeIdentity,
            clearAll: state.clearAll,
        })),
    );
};

export const useInboxPageStore = () => {
    return useIdentityStore(
        useShallow((state) => ({
            identities: state.identities,
            activeIndex: state.activeIndex,
            currentAddress: state.identities[state.activeIndex]?.address ?? '',
            setActiveIndex: state.setActiveIndex,
            removeIdentity: state.removeIdentity,
        })),
    );
};

export const useAddInboxStore = () => {
    return useIdentityStore(
        useShallow((state) => ({
            mnemonic: state.mnemonic,
            identities: state.identities,
            addIdentity: state.addIdentity,
            removeIdentity: state.removeIdentity,
            setActiveIndex: state.setActiveIndex,
        })),
    );
};

export const useIdentityStoreSelectors = <T>(
    selector: (state: ReturnType<typeof useIdentityStore.getState>) => T,
) => {
    return useIdentityStore(useShallow(selector));
};
