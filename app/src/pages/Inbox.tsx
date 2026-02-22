import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { InboxHeader } from '@/components/base/Header/InboxHeader';
import { AccountSidebar, type Account } from '@/components/base/AccountSidebar';
import { InboxList, mockEmails, type Email } from '@/components/base/InboxList';
import { MessagePanel } from '@/components/base/MessagePanel';
import { useIdentityStore } from '@/lib/stores/identityStore';
import { useAddInbox } from '@/hooks/useAddInbox';

const Inbox = () => {
    const identities = useIdentityStore((s) => s.identities);
    const activeIndex = useIdentityStore((s) => s.activeIndex);
    const setActiveIndex = useIdentityStore((s) => s.setActiveIndex);
    const removeIdentity = useIdentityStore((s) => s.removeIdentity);
    const { addInbox, isAddDisabled } = useAddInbox();

    const [sidebarExpanded, setSidebarExpanded] = useState(false);
    const [selectedEmail, setSelectedEmail] = useState<Email | null>(null);
    const [emails, setEmails] = useState(mockEmails);

    if (identities.length === 0) {
        return <Navigate to='/' replace />;
    }

    const currentAddress = identities[activeIndex]?.address ?? '';

    const accounts: Account[] = identities.map((identity) => ({
        id: identity.address,
        address: identity.address,
        messageCount: 0,
    }));

    const handleToggleSidebar = () => setSidebarExpanded((prev) => !prev);

    const handleDeleteAccount = (index: number) => {
        removeIdentity(index);
    };

    const handleDeleteEmail = (emailId: string) => {
        setEmails((prev) => prev.filter((email) => email.id !== emailId));
        if (selectedEmail?.id === emailId) {
            setSelectedEmail(null);
        }
    };

    return (
        <div className='w-full h-dvh flex flex-col'>
            {/* Invisible, SEO purpose only */}
            <h1 className='sr-only'>Stateless temporary email.</h1>

            <InboxHeader />
            <div className='flex-1 flex overflow-hidden'>
                <AccountSidebar
                    accounts={accounts}
                    selectedIndex={activeIndex}
                    currentAddress={currentAddress}
                    isExpanded={sidebarExpanded}
                    onToggle={handleToggleSidebar}
                    onSelectAccount={setActiveIndex}
                    onAddAccount={addInbox}
                    isAddDisabled={isAddDisabled}
                    onDeleteAccount={handleDeleteAccount}
                />
                <InboxList
                    emails={emails}
                    selectedEmailId={selectedEmail?.id ?? null}
                    onSelectEmail={setSelectedEmail}
                    onDeleteEmail={handleDeleteEmail}
                    onToggleSidebar={handleToggleSidebar}
                />
                <MessagePanel
                    email={selectedEmail}
                    onDeleteEmail={() => {
                        if (selectedEmail) {
                            handleDeleteEmail(selectedEmail.id);
                        }
                    }}
                />
            </div>
        </div>
    );
};

export default Inbox;
