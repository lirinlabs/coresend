import { forwardRef } from "react";
import Typography from "../Typography/typography";

interface SeedSlotProps {
    index: number;
    value: string;
    onChange: (value: string) => void;
    onKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void;
}

const SeedSlot = forwardRef<HTMLInputElement, SeedSlotProps>(
    ({ index, value, onChange, onKeyDown }, ref) => {
        const slotNumber = String(index + 1).padStart(2, "0");

        return (
            <div className="border border-foreground bg-background px-3 py-2 transition-colors focus-within:bg-secondary flex items-baseline gap-2">
                <Typography
                    color="muted"
                    font="mono"
                    text="xs"
                    className="leading-tight"
                >
                    {slotNumber}.
                </Typography>
                <input
                    ref={ref}
                    type="text"
                    value={value}
                    onChange={(e) =>
                        onChange(e.target.value.toLowerCase().trim())
                    }
                    onKeyDown={onKeyDown}
                    className="text-sm flex-1 bg-transparent outline-none font-mono font-normal text-foreground placeholder:text-muted-foreground placeholder:text-sm placeholder:font-mono placeholder:font-light leading-tight mt-0.5"
                    placeholder="word"
                    autoComplete="off"
                    spellCheck={false}
                />
            </div>
        );
    }
);

SeedSlot.displayName = "SeedSlot";

export default SeedSlot;
