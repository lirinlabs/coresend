import TrashIcon from "@/components/ui/trash-icon"
import Typography from "../Typography/typography"
import { cn } from "@/lib/utils"
import type { Email } from "./types"

interface InboxListItemProps {
  email: Email
  isSelected: boolean
  onSelect: (email: Email | null) => void
  onDelete: () => void
}

export const InboxListItem = ({
  email,
  isSelected,
  onSelect,
  onDelete,
}: InboxListItemProps) => {
  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    onDelete()
  }

  const handleSelect = () => {
    onSelect(email)
  }

  return (
    <button
      type="button"
      onClick={handleSelect}
      className={cn(
        "group w-full text-left px-4 py-3 border-b border-border cursor-pointer transition-colors",
        isSelected
          ? "bg-secondary border-l-2 border-l-primary"
          : "hover:bg-secondary/50"
      )}
    >
      <div className="flex items-start justify-between gap-2">
        <Typography
          as="span"
          text="sm"
          weight="medium"
          color="foreground"
          className="truncate flex-1"
        >
          {email.from}
        </Typography>
        <button
          type="button"
          className="opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
          onClick={handleDelete}
        >
          <TrashIcon size={16} dangerHover className="text-muted-foreground" />
        </button>
      </div>
      <Typography
        as="p"
        text="sm"
        color="muted"
        className="truncate mt-0.5"
      >
        {email.subject}
      </Typography>
      <Typography as="p" text="xs" font="mono" color="muted" className="mt-1">
        TTL: <span className="text-primary">{email.ttl}</span>
      </Typography>
    </button>
  )
}
