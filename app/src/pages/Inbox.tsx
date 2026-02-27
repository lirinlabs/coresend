import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { InboxHeader } from '@/components/base/Header/InboxHeader';
import { AccountSidebar, type Account } from '@/components/base/AccountSidebar';
import { InboxList, type Email } from '@/components/base/InboxList';
import { MessagePanel } from '@/components/base/MessagePanel';
import { useInboxPageStore } from '@/lib/stores/identityStore.selectors';
import { useAddInbox } from '@/hooks/useAddInbox';

const Inbox = () => {
    const {
        identities,
        activeIndex,
        currentAddress,
        setActiveIndex,
        removeIdentity,
    } = useInboxPageStore();
    const { addInbox, isAddDisabled } = useAddInbox();

    const [sidebarExpanded, setSidebarExpanded] = useState(false);
    const [emailsByAccount, setEmailsByAccount] = useState<
        Record<string, Email[]>
    >({});
    const [selectedByAccount, setSelectedByAccount] = useState<
        Record<string, string | null>
    >({});

    const emails = emailsByAccount[currentAddress] ?? [];

    const selectedEmailId = selectedByAccount[currentAddress] ?? null;
    const selectedEmail = emails.find((e) => e.id === selectedEmailId) ?? null;

    if (identities.length === 0) {
        return <Navigate to='/' replace />;
    }

    const accounts: Account[] = identities.map((identity) => ({
        id: identity.address,
        address: identity.address,
        messageCount: 0,
    }));

    const handleToggleSidebar = () => setSidebarExpanded((prev) => !prev);

    const handleDeleteAccount = (index: number) => {
        removeIdentity(index);
    };

    const setEmails = (updater: (prev: Email[]) => Email[]) => {
        setEmailsByAccount((prev) => ({
            ...prev,
            [currentAddress]: updater(prev[currentAddress] ?? []),
        }));
    };

    const setSelectedEmail = (email: Email | null) => {
        setSelectedByAccount((prev) => ({
            ...prev,
            [currentAddress]: email?.id ?? null,
        }));
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
