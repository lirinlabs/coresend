import Typography from "../Typography/typography";

export const Logo = () => {
    return (
        <Typography
            weight="semibold"
            text="sm"
            tracking="tight"
            font="mono"
            color="foreground"
            as="span"
            transform="uppercase"
            className="leading-none"
        >
            CORESEND
        </Typography>
    );
};
