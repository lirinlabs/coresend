import { useNavigate } from "react-router-dom";
import Button from "@/components/base/Button/Button";
import Typography from "@/components/base/Typography/typography";
import { Footer } from "@/components/base/Footer/Footer";
import { Header } from "@/components/base/Header/Header";

const Landing = () => {
    const navigate = useNavigate();
    return (
        <div className="w-full h-dvh flex flex-col">
            {/* Invisible, SEO purpose only */}
            <h1 className="sr-only">Stateless temporary email.</h1>

            <Header />
            <div className="max-w-7xl mx-auto w-full flex-1 flex flex-col justify-center items-center">
                <div className="flex flex-col items-center gap-4">
                    <div className="flex flex-col">
                        <Typography
                            weight="extrabold"
                            align="center"
                            text="6xl"
                        >
                            Your Inbox.
                        </Typography>
                        <Typography
                            weight="extrabold"
                            align="center"
                            text="6xl"
                        >
                            Derived from Math.
                        </Typography>
                    </div>
                    <div className="flex flex-col">
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
                        className="mt-4"
                        onClick={() => {
                            navigate("/gateway");
                        }}
                    >
                        ENTER GATEWAY
                    </Button>
                </div>
            </div>
            <Footer />
        </div>
    );
};

export default Landing;
