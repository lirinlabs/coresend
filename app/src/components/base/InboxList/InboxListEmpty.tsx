import Typography from "../Typography/typography"

export const InboxListEmpty = () => {
  return (
    <div className="flex-1 flex items-center justify-center p-4">
      <Typography text="xs" font="mono" color="muted">
        [ NO_INBOUND_DATA ]
      </Typography>
    </div>
  )
}
