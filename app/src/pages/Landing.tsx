import Button from "@/components/base/Button/Button";
import Typography from "@/components/base/Typography/typography";

const Landing = () => {
    return (
        <div className="w-full h-dvh flex flex-col justify-center items-center gap-4">
            <h1 className="sr-only">Stateless temporary email.</h1>

            <div className="flex flex-col">
                <Typography weight="extrabold" align="center" text="6xl">
                    Your Inbox.
                </Typography>
                <Typography weight="extrabold" align="center" text="6xl">
                    Derived from Math.
                </Typography>
            </div>
            <div className="flex flex-col gap-0">
                <Typography
                    font="mono"
                    weight="normal"
                    align="center"
                    text="sm"
                    color="muted"
                >
                    Powered by BIP39 deterministic logic.
                </Typography>
                <Typography
                    font="mono"
                    weight="normal"
                    align="center"
                    text="sm"
                    color="muted"
                >
                    24h TTL. No database. Inbound only.{" "}
                </Typography>
            </div>
            <Button
                variant="primary"
                size="md"
                onClick={() => {
                    window.location.href = "/app";
                }}
            >
                ENTER GATEWAY
            </Button>
        </div>
    );
};

export default Landing;
