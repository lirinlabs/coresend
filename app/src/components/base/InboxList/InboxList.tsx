import type { InboxListProps } from "./types"
import { InboxListHeader } from "./InboxListHeader"
import { InboxListItem } from "./InboxListItem"
import { InboxListEmpty } from "./InboxListEmpty"

export const InboxList = ({
  emails,
  selectedEmailId,
  onSelectEmail,
  onDeleteEmail,
  onToggleSidebar,
}: InboxListProps) => {
  return (
    <div className="w-72 border-r border-border flex flex-col h-full">
      <InboxListHeader onToggleSidebar={onToggleSidebar} />
      <div className="flex-1 overflow-y-auto">
        {emails.length === 0 ? (
          <InboxListEmpty />
        ) : (
          emails.map((email) => (
            <InboxListItem
              key={email.id}
              email={email}
              isSelected={selectedEmailId === email.id}
              onSelect={() => onSelectEmail(email)}
              onDelete={() => onDeleteEmail(email.id)}
            />
          ))
        )}
      </div>
    </div>
  )
}
