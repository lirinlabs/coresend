import { motion } from 'motion/react';
import type { AccountSidebarProps } from './types';
import { AccountAvatar } from './AccountAvatar';
import { AccountIndicator } from './AccountIndicator';
import { AccountListItem } from './AccountListItem';
import { AddAccountButton } from './AddAccountButton';
import Typography from '../Typography/typography';

export const AccountSidebar = ({
    accounts,
    selectedIndex,
    currentAddress,
    isExpanded,
    onToggle,
    onSelectAccount,
    onAddAccount,
    isAddDisabled = false,
}: AccountSidebarProps) => {
    const totalUnread = accounts.reduce(
        (sum, acc) => sum + acc.messageCount,
        0,
    );

    return (
        <motion.aside
            initial={false}
            animate={{ width: isExpanded ? 242 : 56 }}
            transition={{ duration: 0.2, ease: 'easeInOut' }}
            className='hidden md:flex border-r border-border flex-col shrink-0 overflow-hidden'
        >
            <div className='flex-1 overflow-y-auto py-2'>
                {/* Avatar section */}
                <div className='flex flex-col items-center gap-1 pb-2 border-b border-border mx-2 mb-2'>
                    <AccountAvatar
                        address={currentAddress}
                        onClick={onToggle}
                        showTooltip={!isExpanded}
                    />
                </div>

                {/* Inbox label - only visible when expanded */}
                {isExpanded && (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        transition={{ delay: 0.1 }}
                        className='px-3 py-2 mx-2 mb-2'
                    >
                        <Typography
                            as='p'
                            text='sm'
                            weight='medium'
                            color='foreground'
                            tracking='tight'
                        >
                            Inbox
                        </Typography>
                        <Typography
                            as='p'
                            text='xs'
                            color='muted'
                            weight='light'
                            tracking='tight'
                        >
                            {totalUnread} unread
                        </Typography>
                    </motion.div>
                )}

                {/* Account list */}
                {isExpanded ? (
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        transition={{ delay: 0.1 }}
                        className='flex flex-col px-2'
                    >
                        {accounts.map((account, index) => (
                            <AccountListItem
                                key={account.id}
                                account={account}
                                isSelected={selectedIndex === index}
                                onClick={() => onSelectAccount(index)}
                            />
                        ))}
                    </motion.div>
                ) : (
                    <div className='flex flex-col items-center gap-2 mt-2'>
                        {accounts.map((account, index) => (
                            <AccountIndicator
                                key={account.id}
                                index={index}
                                address={account.address}
                                isSelected={selectedIndex === index}
                                onClick={() => onSelectAccount(index)}
                            />
                        ))}
                    </div>
                )}
            </div>

            {/* Add account button */}
            <div className='p-2 border-t border-border flex justify-center'>
                <AddAccountButton
                    onClick={onAddAccount}
                    disabled={isAddDisabled}
                />
            </div>
        </motion.aside>
    );
};
