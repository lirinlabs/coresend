import { useShallow } from 'zustand/react/shallow';
import { useIdentityStore } from './identityStore';

export const useInboxHeaderStore = () => {
    return useIdentityStore(
        useShallow((state) => ({
            identities: state.identities,
            activeIndex: state.activeIndex,
            removeIdentity: state.removeIdentity,
            clearAll: state.clearAll,
            getActiveIdentity: state.getActiveIdentity,
        })),
    );
};

export const useInboxPageStore = () => {
    return useIdentityStore(
        useShallow((state) => ({
            identities: state.identities,
            activeIndex: state.activeIndex,
            setActiveIndex: state.setActiveIndex,
            removeIdentity: state.removeIdentity,
            hasIdentities: state.hasIdentities,
            getActiveAddress: state.getActiveAddress,
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
