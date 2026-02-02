import { useState } from "react";
import { InboxHeader } from "@/components/base/Header/InboxHeader";
import {
    AccountSidebar,
    mockAccounts,
} from "@/components/base/AccountSidebar";
import { InboxList, mockEmails, type Email } from "@/components/base/InboxList";

const Inbox = () => {
    const [selectedAccount, setSelectedAccount] = useState(0);
    const [sidebarExpanded, setSidebarExpanded] = useState(false);
    const [selectedEmail, setSelectedEmail] = useState<Email | null>(null);
    const [emails, setEmails] = useState(mockEmails);

    const currentAddress = mockAccounts[selectedAccount]?.address ?? "";

    const handleToggleSidebar = () => setSidebarExpanded((prev) => !prev);

    const handleDeleteEmail = (emailId: string) => {
        setEmails((prev) => prev.filter((email) => email.id !== emailId));
        if (selectedEmail?.id === emailId) {
            setSelectedEmail(null);
        }
    };

    return (
        <div className="w-full h-dvh flex flex-col">
            {/* Invisible, SEO purpose only */}
            <h1 className="sr-only">Stateless temporary email.</h1>

            <InboxHeader />
            <div className="flex-1 flex overflow-hidden">
                <AccountSidebar
                    accounts={mockAccounts}
                    selectedIndex={selectedAccount}
                    currentAddress={currentAddress}
                    isExpanded={sidebarExpanded}
                    onToggle={handleToggleSidebar}
                    onSelectAccount={setSelectedAccount}
                    onAddAccount={() => console.log("Add account")}
                />
                <InboxList
                    emails={emails}
                    selectedEmailId={selectedEmail?.id ?? null}
                    onSelectEmail={setSelectedEmail}
                    onDeleteEmail={handleDeleteEmail}
                    onToggleSidebar={handleToggleSidebar}
                />
            </div>
        </div>
    );
};

export default Inbox;
