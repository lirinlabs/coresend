export interface Account {
  id: string
  address: string
  messageCount: number
}

export interface AccountSidebarProps {
  accounts: Account[]
  selectedIndex: number
  currentAddress: string
  onSelectAccount: (index: number) => void
  onAddAccount: () => void
}
