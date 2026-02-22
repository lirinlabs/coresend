import { useIdentityStore } from '@/lib/stores/identityStore';
import { deriveIdentityFromMnemonic } from '@/lib/crypto/deriveIdentityFromMnemonic';
import { useRegisterAddress } from '@/api/generated';

const MAX_INBOXES = 10;

export const useAddInbox = () => {
    const mnemonic = useIdentityStore((state) => state.mnemonic);
    const identities = useIdentityStore((state) => state.identities);
    const addIdentity = useIdentityStore((state) => state.addIdentity);
    const setActiveIndex = useIdentityStore((state) => state.setActiveIndex);

    const registerMutation = useRegisterAddress();

    const canAddInbox = !!mnemonic && identities.length < MAX_INBOXES;

    const addInbox = async () => {
        if (!canAddInbox || !mnemonic) {
            return;
        }

        const nextIndex = identities.length;
        const identity = deriveIdentityFromMnemonic(mnemonic, nextIndex);

        addIdentity(identity);
        setActiveIndex(nextIndex);

        await registerMutation.mutateAsync({
            address: identity.address,
        });
    };

    return {
        addInbox,
        canAddInbox,
        isRegistering: registerMutation.isPending,
    };
};
