import Typography from "../Typography/typography"
import type { Email } from "../InboxList/types"

interface MessagePanelContentProps {
  email: Email
}

export const MessagePanelContent = ({ email }: MessagePanelContentProps) => {
  return (
    <div className="flex-1 overflow-y-auto p-6">
      {/* Subject */}
      <Typography as="h2" text="xl" weight="semibold" className="mb-4">
        {email.subject}
      </Typography>

      {/* Metadata */}
      <div className="space-y-1 mb-4">
        <Typography as="p" text="sm" font="mono">
          <span className="text-muted-foreground">From: </span>
          {email.from}
        </Typography>
        <Typography as="p" text="sm" font="mono">
          <span className="text-muted-foreground">Time: </span>
          {email.timestamp.toISOString()}
        </Typography>
        <Typography as="p" text="sm" font="mono">
          <span className="text-muted-foreground">TTL: </span>
          <span className="text-primary">{email.ttl}</span>
        </Typography>
      </div>

      {/* Divider */}
      <div className="border-b border-border mb-6" />

      {/* Body */}
      <Typography
        as="div"
        text="sm"
        color="foreground"
        className="whitespace-pre-wrap leading-relaxed"
      >
        {email.body}
      </Typography>
    </div>
  )
}
