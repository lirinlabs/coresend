import { Plus } from '@phosphor-icons/react';
import {
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from '@/components/ui/tooltip';

interface AddAccountButtonProps {
    onClick: () => void;
    disabled?: boolean;
}

export const AddAccountButton = ({
    onClick,
    disabled = false,
}: AddAccountButtonProps) => {
    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <button
                    type='button'
                    onClick={onClick}
                    disabled={disabled}
                    className='w-8 h-8 border border-dashed border-border rounded flex items-center justify-center hover:bg-secondary hover:border-solid transition-colors disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-transparent disabled:hover:border-dashed'
                >
                    <Plus
                        weight='bold'
                        className='w-4 h-4 text-muted-foreground'
                    />
                </button>
            </TooltipTrigger>
            <TooltipContent side='right'>
                {disabled ? 'Maximum 10 inboxes' : 'Derive new address'}
            </TooltipContent>
        </Tooltip>
    );
};
