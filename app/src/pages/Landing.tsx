import Typography from "@/components/base/Typography/typography";

const Landing = () => {
    return (
        <div className="w-full h-dvh flex flex-col justify-center items-center gap-4">
            <h1 className="sr-only">Stateless Inbox</h1>

            <div className="bg-accent">
                <Typography color="muted" text="xs" font="mono">
                    Ephemeral • Stateless • Private
                </Typography>
            </div>
            <div className="flex flex-col">
                <Typography weight="extrabold" align="center" text="6xl">
                    Your Inbox.
                </Typography>
                <Typography weight="extrabold" align="center" text="6xl">
                    Derived from Math.
                </Typography>
                <Typography
                    weight="extrabold"
                    align="center"
                    text="6xl"
                    color="muted"
                >
                    Not Stored.
                </Typography>
            </div>
            <Typography weight="normal" align="center" text="lg" color="muted">
                A temporary email that works like a crypto wallet. No accounts.
                No passwords. Just a 12-word seed phrase that unlocks your
                ephemeral inbox.
            </Typography>
        </div>
    );
};

export default Landing;
