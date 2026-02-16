import TrashIcon from '@/components/ui/trash-icon';
import Typography from '../Typography/typography';

interface MessagePanelHeaderProps {
    showDelete: boolean;
    onDelete: () => void;
}

export const MessagePanelHeader = ({
    showDelete,
    onDelete,
}: MessagePanelHeaderProps) => {
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
                Message
            </Typography>
            {showDelete && (
                <button
                    type='button'
                    onClick={onDelete}
                    className='p-1 text-muted-foreground hover:text-foreground transition-colors'
                >
                    <TrashIcon size={16} dangerHover />
                </button>
            )}
        </div>
    );
};
