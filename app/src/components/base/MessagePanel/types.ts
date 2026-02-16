import type { Email } from '../InboxList/types';

export interface MessagePanelProps {
    email: Email | null;
    onDeleteEmail: () => void;
}
