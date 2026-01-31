import Typography from "@/components/base/Typography/typography";

const Landing = () => {
    return (
        <div className="w-full h-dvh flex flex-col justify-center items-center gap-4">
            <div className="bg-accent">
                <Typography color="muted" text="sm" font="mono">
                    Ephemeral • Stateless • Private
                </Typography>
            </div>
            <div className="flex flex-col">
                <Typography text="9xl">Your Inbox.</Typography>
                <Typography text="9xl">Derived from Math.</Typography>
                <Typography color="muted" text="9xl">
                    Not Stored.
                </Typography>
              
            </div>
            <Typography>
                A temporary email that works like a crypto wallet. No accounts.
                No passwords. Just a 12-word seed phrase that unlocks your
                ephemeral inbox.
            </Typography>
        </div>
    );
};

export default Landing;
