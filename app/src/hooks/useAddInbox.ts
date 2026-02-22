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
        if (!canAddInbox || !mnemonic || registerMutation.isPending) {
            return;
        }

        const nextIndex = identities.length;
        const identity = deriveIdentityFromMnemonic(mnemonic, nextIndex);

        addIdentity(identity);
        setActiveIndex(nextIndex);

        try {
            await registerMutation.mutateAsync({ address: identity.address });
            toast.success(`Inbox ${nextIndex + 1} created`);
        } catch {
            removeIdentity(nextIndex);
            setActiveIndex(Math.max(0, identities.length - 1));
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
