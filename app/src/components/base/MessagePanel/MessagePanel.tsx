import type { MessagePanelProps } from './types';
import { MessagePanelHeader } from './MessagePanelHeader';
import { MessagePanelContent } from './MessagePanelContent';
import { MessagePanelEmpty } from './MessagePanelEmpty';

export const MessagePanel = ({ email, onDeleteEmail }: MessagePanelProps) => {
    return (
        <main className='flex-1 flex flex-col h-full overflow-hidden'>
            <MessagePanelHeader showDelete={!!email} onDelete={onDeleteEmail} />
            {email ? (
                <MessagePanelContent email={email} />
            ) : (
                <MessagePanelEmpty />
            )}
        </main>
    );
};
