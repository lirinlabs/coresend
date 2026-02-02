import type { AccountSidebarProps } from "./types"
import { AccountAvatar } from "./AccountAvatar"
import { AccountIndicator } from "./AccountIndicator"
import { AddAccountButton } from "./AddAccountButton"

export const AccountSidebar = ({
  accounts,
  selectedIndex,
  currentAddress,
  onSelectAccount,
  onAddAccount,
}: AccountSidebarProps) => {
  return (
    <aside className="hidden md:flex w-14 border-r border-border flex-col shrink-0">
      <div className="flex-1 overflow-y-auto py-2">
        {/* Current account avatar */}
        <div className="flex flex-col items-center gap-1 pb-2 border-b border-border mx-2 mb-2">
          <AccountAvatar address={currentAddress} />
        </div>

        {/* Account list */}
        <div className="flex flex-col items-center gap-2 mt-2">
          {accounts.map((account, index) => (
            <AccountIndicator
              key={account.id}
              index={index}
              address={account.address}
              isSelected={selectedIndex === index}
              onClick={() => onSelectAccount(index)}
            />
          ))}
        </div>
      </div>

      {/* Add account button */}
      <div className="p-2 border-t border-border flex justify-center">
        <AddAccountButton onClick={onAddAccount} />
      </div>
    </aside>
  )
}
