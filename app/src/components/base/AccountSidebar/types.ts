export interface Account {
    id: string;
    address: string;
    messageCount: number;
}

export interface AccountSidebarProps {
    accounts: Account[];
    selectedIndex: number;
    currentAddress: string;
    isExpanded: boolean;
    onToggle: () => void;
    onSelectAccount: (index: number) => void;
    onAddAccount: () => void;
    isAddDisabled?: boolean;
    onDeleteAccount?: (index: number) => void;
}
