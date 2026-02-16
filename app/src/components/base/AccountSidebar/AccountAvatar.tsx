import {
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

interface AccountAvatarProps {
    address: string;
    className?: string;
    onClick?: () => void;
    showTooltip?: boolean;
}

/**
 * Generates a deterministic 3x3 pattern from the first 9 characters of an address.
 * Each character's char code determines if the cell is filled (odd = filled).
 */
function generatePattern(address: string): boolean[] {
    const chars = address.replace('@', '').slice(0, 9);
    return Array.from({ length: 9 }, (_, i) => {
        const charCode = chars.charCodeAt(i) || 0;
        return charCode % 2 === 1;
    });
}

export const AccountAvatar = ({
    address,
    className,
    onClick,
    showTooltip = true,
}: AccountAvatarProps) => {
    const pattern = generatePattern(address);

    const avatarContent = (
        <button
            type='button'
            onClick={onClick}
            className={cn(
                'w-10 h-10 border border-border rounded-md flex items-center justify-center cursor-pointer hover:bg-secondary transition-colors',
                className,
            )}
        >
            <div className='grid grid-cols-3 gap-0.5'>
                {pattern.map((filled, i) => (
                    <div
                        key={`cell-${i}-${filled}`}
                        className={cn(
                            'w-1.5 h-1.5',
                            filled ? 'bg-foreground' : 'bg-transparent',
                        )}
                    />
                ))}
            </div>
        </button>
    );

    if (!showTooltip) {
        return avatarContent;
    }

    return (
        <Tooltip>
            <TooltipTrigger asChild>{avatarContent}</TooltipTrigger>
            <TooltipContent side='right'>Current: {address}</TooltipContent>
        </Tooltip>
    );
};
