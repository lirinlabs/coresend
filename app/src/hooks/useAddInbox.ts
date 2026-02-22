import { useIdentityStore } from '@/lib/stores/identityStore';
import { deriveIdentityFromMnemonic } from '@/lib/crypto/deriveIdentityFromMnemonic';
import { useRegisterAddress } from '@/api/generated';
import { toast } from 'sonner';

const MAX_INBOXES = 10;

export const useAddInbox = () => {
    const mnemonic = useIdentityStore((state) => state.mnemonic);
    const identities = useIdentityStore((state) => state.identities);
    const addIdentity = useIdentityStore((state) => state.addIdentity);
    const removeIdentity = useIdentityStore((state) => state.removeIdentity);
    const setActiveIndex = useIdentityStore((state) => state.setActiveIndex);

    const registerMutation = useRegisterAddress();

    const canAddInbox = !!mnemonic && identities.length < MAX_INBOXES;
    const isAddDisabled = !canAddInbox || registerMutation.isPending;

    const addInbox = async () => {
        if (!canAddInbox || !mnemonic || registerMutation.isPending) return;

        const state = useIdentityStore.getState();
        const existingIndexes = state.identities
            .map((i) => i.index)
            .sort((a, b) => a - b);
        let nextDerivationIndex = 0;
        for (const idx of existingIndexes) {
            if (idx === nextDerivationIndex) nextDerivationIndex++;
            else if (idx > nextDerivationIndex) break;
        }

        const identity = deriveIdentityFromMnemonic(
            mnemonic,
            nextDerivationIndex,
        );
        const prevActiveIndex = state.activeIndex;
        const insertPosition = state.identities.length;

        addIdentity(identity);
        setActiveIndex(insertPosition);

        try {
            await registerMutation.mutateAsync({ address: identity.address });
            toast.success(
                'New inbox created, this is your new email: ' +
                    identity.address,
            );
        } catch {
            const currentState = useIdentityStore.getState();
            const addedIndex = currentState.identities.findIndex(
                (i) => i.address === identity.address,
            );
            if (addedIndex !== -1) removeIdentity(addedIndex);

            const rollbackState = useIdentityStore.getState();
            if (rollbackState.identities.length > 0) {
                setActiveIndex(
                    Math.min(
                        prevActiveIndex,
                        rollbackState.identities.length - 1,
                    ),
                );
            }
            toast.error('Failed to register inbox. Please try again.');
        }
    };

    return {
        addInbox,
        canAddInbox,
        isAddDisabled,
        isRegistering: registerMutation.isPending,
    };
};
