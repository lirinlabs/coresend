import { useState } from "react";
import { InboxHeader } from "@/components/base/Header/InboxHeader";
import {
    AccountSidebar,
    mockAccounts,
} from "@/components/base/AccountSidebar";

const Inbox = () => {
    const [selectedAccount, setSelectedAccount] = useState(0);
    const currentAddress = mockAccounts[selectedAccount]?.address ?? "";

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
                    onSelectAccount={setSelectedAccount}
                    onAddAccount={() => console.log("Add account")}
                />
            </div>
        </div>
    );
};

export default Inbox;
