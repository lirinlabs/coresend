export interface Email {
  id: string
  from: string
  subject: string
  body: string
  timestamp: Date
  ttl: string
}

export interface InboxListProps {
  emails: Email[]
  selectedEmailId: string | null
  onSelectEmail: (email: Email) => void
  onDeleteEmail: (emailId: string) => void
  onToggleSidebar: () => void
}
