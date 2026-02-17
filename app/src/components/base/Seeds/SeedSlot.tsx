import { forwardRef, useState } from 'react';
import Typography from '../Typography/typography';

interface SeedSlotProps {
    index: number;
    value: string;
    onChange: (value: string) => void;
    onKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void;
}

const SeedSlot = forwardRef<HTMLInputElement, SeedSlotProps>(
    ({ index, value, onChange, onKeyDown }, ref) => {
        const slotNumber = String(index + 1).padStart(2, '0');
        const [isFocused, setIsFocused] = useState(false);
        return (
            <div className='border border-foreground bg-background px-3 py-2 transition-colors focus-within:bg-secondary flex items-baseline gap-2'>
                <Typography
                    color='muted'
                    font='mono'
                    text='xs'
                    className='leading-normal'
                >
                    {slotNumber}.
                </Typography>
                <input
                    ref={ref}
                    type={isFocused ? 'text' : 'password'}
                    autoCorrect='off'
                    autoCapitalize='off'
                    name='field-1'
                    onFocus={() => setIsFocused(true)}
                    onBlur={() => setIsFocused(false)}
                    value={value}
                    onChange={(e) =>
                        onChange(e.target.value.toLowerCase().trim())
                    }
                    onKeyDown={onKeyDown}
                    className='text-sm flex-1 bg-transparent outline-none font-mono font-normal text-foreground placeholder:text-muted-foreground placeholder:text-sm placeholder:font-mono placeholder:font-light leading-tight mt-0.5'
                    placeholder='word'
                    autoComplete='off'
                    spellCheck={false}
                />
            </div>
        );
    },
);

SeedSlot.displayName = 'SeedSlot';

export default SeedSlot;
