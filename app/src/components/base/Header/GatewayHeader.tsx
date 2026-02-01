import { useNavigate } from "react-router-dom";
import Button from "../Button/Button";
import { ModeToggle } from "../ModeToggle/ModeToggle";
import Typography from "../Typography/typography";
import { ArrowLeftIcon } from "@phosphor-icons/react";

export const GatewayHeader = () => {
    const navigate = useNavigate();
    return (
        <header className="w-full py-2 border-b border-border">
            <div className="max-w-7xl mx-auto flex items-center justify-between">
                <Button
                    variant="ghost"
                    size="sm"
                    className="hover:text-primary cursor-pointer"
                    onClick={() => {
                        navigate("/");
                    }}
                    leftIcon={<ArrowLeftIcon size={12} weight="regular" />}
                >
                    <Typography
                        weight="medium"
                        text="xs"
                        tracking="tight"
                        font="mono"
                        transform="uppercase"
                        className="leading-none"
                    >
                        RETURN_TO_ORIGIN
                    </Typography>
                </Button>
                <ModeToggle />
            </div>
        </header>
    );
};
