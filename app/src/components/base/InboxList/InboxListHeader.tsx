import { CaretRightIcon } from '@phosphor-icons/react';
import Typography from '../Typography/typography';

interface InboxListHeaderProps {
    onToggleSidebar: () => void;
}

export const InboxListHeader = ({ onToggleSidebar }: InboxListHeaderProps) => {
    return (
        <div className='h-12 px-4 flex items-center justify-between border-b border-border shrink-0'>
            <Typography
                as='span'
                text='xs'
                font='mono'
                weight='medium'
                tracking='normal'
                color='muted'
                transform='uppercase'
            >
                Inbox
            </Typography>
            <button
                type='button'
                onClick={onToggleSidebar}
                className='p-1 text-muted-foreground hover:text-foreground transition-colors'
            >
                <CaretRightIcon weight='bold' className='w-4 h-4' />
            </button>
        </div>
    );
};
