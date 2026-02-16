import { EnvelopeSimple } from '@phosphor-icons/react';
import Typography from '../Typography/typography';
import { cn } from '@/lib/utils';
import type { Account } from './types';

interface AccountListItemProps {
    account: Account;
    isSelected: boolean;
    onClick: () => void;
}

export const AccountListItem = ({
    account,
    isSelected,
    onClick,
}: AccountListItemProps) => {
    return (
        <button
            type='button'
            onClick={onClick}
            className={cn(
                'w-full px-3 py-2 flex items-center gap-2 cursor-pointer transition-colors',
                isSelected ? 'bg-secondary' : 'hover:bg-secondary/50',
            )}
        >
            <EnvelopeSimple
                weight='regular'
                className='w-4 h-4 shrink-0 text-muted-foreground'
            />
            <Typography
                as='span'
                text='xs'
                font='mono'
                color='foreground'
                className='flex-1 text-left truncate'
            >
                {account.address}
            </Typography>
            {account.messageCount > 0 && (
                <Typography
                    as='span'
                    text='xs'
                    font='mono'
                    color='primary'
                    weight='medium'
                >
                    {account.messageCount}
                </Typography>
            )}
        </button>
    );
};
