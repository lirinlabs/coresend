import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import TrashIcon from '@/components/ui/trash-icon';
import { Logo } from '../Logo/Logo';
import { ModeToggle } from '../ModeToggle/ModeToggle';
import CopyIcon from '@/components/ui/copy-icon';
import Typography from '../Typography/typography';
import { ActionIcon } from '../ActionIcon';
import { SettingsMenu } from '../SettingsMenu/SettingsMenu';
import { ConfirmDialog } from '../ConfirmDialog/ConfirmDialog';
import { copyToClipboard } from '@/lib/utils';
import { useIdentityStore } from '@/lib/stores/identityStore';

export const InboxHeader = () => {
    const [isCopied, setIsCopied] = useState(false);
    const [showLogoutDialog, setShowLogoutDialog] = useState(false);
    const timeoutRef = useRef<number | null>(null);
    const navigate = useNavigate();
    const identities = useIdentityStore((state) => state.identities);
    const activeIndex = useIdentityStore((state) => state.activeIndex);
    const removeIdentity = useIdentityStore((state) => state.removeIdentity);
    const clearAll = useIdentityStore((state) => state.clearAll);
    const identity = identities[activeIndex];

    const getEmailAddress = () => {
        if (identity) {
            const fullAddress = `${identity.address}@coresend.io`;
            const truncated = `${identity.address.slice(0, 4)}...${identity.address.slice(-4)}@coresend.io`;
            return { fullAddress, displayAddress: truncated };
        }
        const fallback = '4df1234567890432C@coresend.io';
        return {
            fullAddress: fallback,
            displayAddress: '4df...432C@coresend.io',
        };
    };

    const handleCopy = async () => {
        const { fullAddress } = getEmailAddress();
        const success = await copyToClipboard(fullAddress);

        if (success) {
            if (timeoutRef.current) {
                clearTimeout(timeoutRef.current);
            }

            setIsCopied(true);

            timeoutRef.current = setTimeout(() => {
                setIsCopied(false);
                timeoutRef.current = null;
            }, 2000) as unknown as number;
        }
    };

    useEffect(() => {
        return () => {
            if (timeoutRef.current) {
                clearTimeout(timeoutRef.current);
            }
        };
    }, []);

    const { displayAddress } = getEmailAddress();

    const handleDeleteInbox = () => {
        if (identities.length === 1) {
            clearAll();
            navigate('/');
            toast.success('Session ended');
        } else {
            const removedIndex = activeIndex;
            removeIdentity(activeIndex);
            toast.success(`Inbox ${removedIndex + 1} deleted`);
        }
    };

    const handleLogout = () => {
        setShowLogoutDialog(true);
    };

    const confirmLogout = () => {
        clearAll();
        navigate('/');
        toast.success('Logged out');
        setShowLogoutDialog(false);
    };

    return (
        <header className='w-full px-3 md:px-6 py-3 border-b border-border'>
            <div className=' mx-auto flex items-center justify-between'>
                <Logo navigate />

                <div className='flex items-center gap-4'>
                    <button
                        onClick={handleCopy}
                        className='flex items-center gap-2 bg-secondary p-1.5 -my-2 pr-2 rounded-xs 
                                   hover:bg-secondary/80 transition-colors cursor-pointer'
                        type='button'
                        aria-label='Copy email address to clipboard'
                    >
                        <Typography
                            color={isCopied ? 'primary' : 'muted'}
                            text='xs'
                            font='mono'
                            weight='semibold'
                            className='transition-colors'
                        >
                            {isCopied ? '✓ Copied!' : displayAddress}
                        </Typography>
                        <CopyIcon className='text-muted-foreground hover:text-primary transition-colors h-3 w-3' />
                    </button>
                    <ModeToggle />

                    <ActionIcon
                        icon={
                            <TrashIcon className='text-muted-foreground hover:text-primary transition-colors h-4 w-4 ' />
                        }
                        tooltip='Delete inbox'
                        title='Delete Inbox'
                        description='This will permanently delete this inbox. If this is your last inbox, you will be returned to the home screen.'
                        actionText='Delete'
                        onAction={handleDeleteInbox}
                    />

                    <SettingsMenu />

                    <Typography
                        text='xs'
                        font='mono'
                        tracking='tight'
                        transform='uppercase'
                        color='muted'
                        className='cursor-pointer hover:text-primary transition-colors'
                        onClick={handleLogout}
                    >
                        LOGOUT
                    </Typography>
                </div>
            </div>

            <ConfirmDialog
                isOpen={showLogoutDialog}
                onClose={() => setShowLogoutDialog(false)}
                onConfirm={confirmLogout}
                title='Logout'
                description='Are you sure you want to logout? This will end your current session and you will need to unlock your inbox again.'
                confirmText='Logout'
            />
        </header>
    );
};
