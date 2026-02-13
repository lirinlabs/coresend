import Typography from "../Typography/typography";

export const MessagePanelEmpty = () => {
    return (
        <div className="flex-1 flex flex-col gap-2.5 items-center justify-center p-4">
            <div className="radar-container mx-auto">
                <div className="radar-sweep" />
                <div className="radar-sweep" />
                <div className="radar-sweep" />
                <div className="radar-dot" />
            </div>
            <Typography text="sm" font="mono" color="muted" tracking="wide">
                [ STATUS: AWAITING_INBOUND_DATA ]
            </Typography>
            <Typography text="xs" font="mono" color="muted">
                // No packet selected.
            </Typography>
        </div>
    );
};
