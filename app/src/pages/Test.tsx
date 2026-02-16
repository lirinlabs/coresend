import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
    Copy,
    Check,
    Plus,
    Mail,
    Trash2,
    Flame,
    AlertTriangle,
    ArrowLeft,
    Menu,
} from 'lucide-react';
import { SettingsMenu } from '@/components/base/SettingsMenu/SettingsMenu';
import {
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from '@/components/ui/sheet';

interface Email {
    id: string;
    from: string;
    subject: string;
    body: string;
    timestamp: Date;
    ttl: string;
}

interface Account {
    address: string;
    emails: Email[];
}

const generateDemoData = (): Account[] => {
    const domains = ['coresend.io', 'coresend.io', 'coresend.io'];
    const prefixes = ['4df...95', '7f2...98', '9a1...42'];

    return prefixes.map((prefix, i) => ({
        address: `${prefix}@${domains[i % domains.length]}`,
        emails:
            i === 0
                ? [
                      {
                          id: '1',
                          from: 'noreply@service.io',
                          subject: 'Verification Code: 847291',
                          body: `Your verification code is: 847291\n\nThis code expires in 10 minutes.\n\nIf you did not request this code, please ignore this email.\n\n—\nAutomated message from Service.io`,
                          timestamp: new Date(Date.now() - 1000 * 60 * 5),
                          ttl: '23h 55m',
                      },
                      {
                          id: '2',
                          from: 'security@platform.dev',
                          subject: 'New login detected from unknown device',
                          body: `A new login was detected on your account.\n\nDevice: Unknown\nLocation: [REDACTED]\nTime: ${new Date().toISOString()}\n\nIf this was not you, please secure your account immediately.\n\n—\nSecurity Team`,
                          timestamp: new Date(Date.now() - 1000 * 60 * 30),
                          ttl: '23h 30m',
                      },
                      {
                          id: '3',
                          from: 'newsletter@crypto.news',
                          subject: '[WEEKLY] Market Update - Week 48',
                          body: `WEEKLY MARKET DIGEST\n\n— BTC: $43,291 (+2.4%)\n— ETH: $2,847 (+1.8%)\n— SOL: $98.42 (+5.2%)\n\nTop Stories:\n1. New regulatory framework proposed\n2. DeFi TVL reaches new highs\n3. Layer 2 adoption accelerates\n\nRead more at crypto.news/weekly`,
                          timestamp: new Date(Date.now() - 1000 * 60 * 120),
                          ttl: '21h 00m',
                      },
                  ]
                : [],
    }));
};

const Inbox = () => {
    const navigate = useNavigate();
    const [accounts, setAccounts] = useState<Account[]>(generateDemoData);
    const [selectedAccount, setSelectedAccount] = useState<number>(0);
    const [selectedEmail, setSelectedEmail] = useState<Email | null>(null);
    const [copied, setCopied] = useState(false);
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
    const [mobileView, setMobileView] = useState<'list' | 'content'>('list');

    // useEffect(() => {
    //     const seed = sessionStorage.getItem("seedPhrase");
    //     if (!seed) {
    //         navigate("/gateway");
    //     }
    // }, [navigate]);

    const currentAccount = accounts[selectedAccount];
    const emails = currentAccount?.emails || [];

    const handleLogout = () => {
        sessionStorage.removeItem('seedPhrase');
        navigate('/');
    };

    const handleCopy = async () => {
        await navigator.clipboard.writeText(currentAccount.address);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const handleDeleteEmail = (emailId: string) => {
        setAccounts((prev) =>
            prev.map((account, i) =>
                i === selectedAccount
                    ? {
                          ...account,
                          emails: account.emails.filter(
                              (e) => e.id !== emailId,
                          ),
                      }
                    : account,
            ),
        );
        if (selectedEmail?.id === emailId) {
            setSelectedEmail(null);
            setMobileView('list');
        }
    };

    const handleWipeInbox = () => {
        setAccounts((prev) =>
            prev.map((account, i) =>
                i === selectedAccount ? { ...account, emails: [] } : account,
            ),
        );
        setSelectedEmail(null);
        setMobileView('list');
    };

    const handleBurnAccount = () => {
        if (accounts.length <= 1) {
            handleWipeInbox();
            return;
        }
        setAccounts((prev) => prev.filter((_, i) => i !== selectedAccount));
        setSelectedAccount(0);
        setSelectedEmail(null);
        setMobileView('list');
    };

    const handleSelectEmail = (email: Email) => {
        setSelectedEmail(email);
        setMobileView('content');
    };

    const handleSelectAccount = (index: number) => {
        setSelectedAccount(index);
        setSelectedEmail(null);
        setMobileView('list');
        setMobileMenuOpen(false);
    };

    return (
        <div className='h-screen bg-background flex flex-col overflow-hidden'>
            {/* Header */}
            <header className='border-b border-border px-3 md:px-6 py-3 flex items-center justify-between shrink-0 gap-2'>
                <div className='flex items-center gap-2 md:gap-6 min-w-0'>
                    {/* Mobile menu button */}
                    <Sheet
                        open={mobileMenuOpen}
                        onOpenChange={setMobileMenuOpen}
                    >
                        <SheetTrigger asChild>
                            <button className='md:hidden p-2 -ml-2 text-muted-foreground hover:text-foreground'>
                                <Menu className='w-5 h-5' />
                            </button>
                        </SheetTrigger>
                        <SheetContent side='left' className='w-72 p-0'>
                            <SheetHeader className='p-4 border-b border-border'>
                                <SheetTitle className='font-mono text-sm'>
                                    Accounts
                                </SheetTitle>
                            </SheetHeader>
                            <div className='flex-1 overflow-y-auto py-2'>
                                {accounts.map((account, index) => (
                                    <div
                                        key={account.address}
                                        onClick={() =>
                                            handleSelectAccount(index)
                                        }
                                        className={`px-4 py-3 cursor-pointer transition-colors flex items-center gap-3 ${
                                            selectedAccount === index
                                                ? 'bg-secondary'
                                                : 'hover:bg-secondary/50'
                                        }`}
                                    >
                                        <Mail className='w-4 h-4 shrink-0 text-muted-foreground' />
                                        <div className='min-w-0 flex-1'>
                                            <div className='font-mono text-sm truncate'>
                                                {account.address}
                                            </div>
                                            <div className='text-xs text-muted-foreground'>
                                                {account.emails.length} messages
                                            </div>
                                        </div>
                                        {account.emails.length > 0 && (
                                            <span className='text-primary text-xs font-medium'>
                                                {account.emails.length}
                                            </span>
                                        )}
                                    </div>
                                ))}
                                <div className='px-4 py-3 border-t border-border mt-2'>
                                    <button className='flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors'>
                                        <Plus className='w-4 h-4' />
                                        <span className='font-mono text-xs'>
                                            DERIVE_NEW_ADDRESS
                                        </span>
                                    </button>
                                </div>
                            </div>
                        </SheetContent>
                    </Sheet>

                    <Link
                        to='/'
                        className='font-mono text-sm font-semibold tracking-tight shrink-0'
                    >
                        CORESEND
                    </Link>
                </div>

                {/* Email address bar - responsive */}
                <div className='flex items-center gap-1 md:gap-2 px-2 md:px-3 py-1.5 bg-secondary rounded-md min-w-0 flex-1 md:flex-none md:max-w-xs'>
                    <span className='font-mono text-xs md:text-sm truncate'>
                        {currentAccount.address}
                    </span>
                    <button
                        onClick={handleCopy}
                        className='p-1 hover:bg-muted rounded transition-colors shrink-0'
                        aria-label='Copy email address'
                    >
                        {copied ? (
                            <Check className='w-3.5 h-3.5 text-primary' />
                        ) : (
                            <Copy className='w-3.5 h-3.5 text-muted-foreground' />
                        )}
                    </button>
                    <div className='hidden sm:flex items-center gap-1 ml-1'>
                        <span className='w-2 h-2 rounded-full bg-primary' />
                        <span className='w-2 h-2 rounded-full bg-amber-400' />
                    </div>
                </div>

                {/* Action buttons - hidden on small mobile */}
                <div className='flex items-center gap-0 md:gap-1 shrink-0'>
                    <AlertDialog>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <AlertDialogTrigger asChild>
                                    <button className='hidden sm:flex p-2 text-muted-foreground hover:text-foreground transition-colors'>
                                        <Trash2 className='w-4 h-4' />
                                    </button>
                                </AlertDialogTrigger>
                            </TooltipTrigger>
                            <TooltipContent>Wipe inbox</TooltipContent>
                        </Tooltip>
                        <AlertDialogContent>
                            <AlertDialogHeader>
                                <AlertDialogTitle>Wipe Inbox</AlertDialogTitle>
                                <AlertDialogDescription>
                                    This will permanently delete all emails in
                                    the current inbox.
                                </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                                <AlertDialogCancel>Cancel</AlertDialogCancel>
                                <AlertDialogAction
                                    onClick={handleWipeInbox}
                                    className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
                                >
                                    Wipe All
                                </AlertDialogAction>
                            </AlertDialogFooter>
                        </AlertDialogContent>
                    </AlertDialog>

                    <AlertDialog>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <AlertDialogTrigger asChild>
                                    <button className='hidden sm:flex p-2 text-muted-foreground hover:text-destructive transition-colors'>
                                        <Flame className='w-4 h-4' />
                                    </button>
                                </AlertDialogTrigger>
                            </TooltipTrigger>
                            <TooltipContent>Burn account</TooltipContent>
                        </Tooltip>
                        <AlertDialogContent>
                            <AlertDialogHeader>
                                <AlertDialogTitle className='flex items-center gap-2'>
                                    <AlertTriangle className='w-5 h-5 text-destructive' />
                                    Burn Account
                                </AlertDialogTitle>
                                <AlertDialogDescription>
                                    This will permanently destroy this email
                                    address and all its data.
                                </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                                <AlertDialogCancel>Cancel</AlertDialogCancel>
                                <AlertDialogAction
                                    onClick={handleBurnAccount}
                                    className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
                                >
                                    Burn Account
                                </AlertDialogAction>
                            </AlertDialogFooter>
                        </AlertDialogContent>
                    </AlertDialog>

                    <SettingsMenu />

                    <button
                        onClick={handleLogout}
                        className='hidden md:block font-mono text-xs text-muted-foreground hover:text-foreground transition-colors ml-2'
                    >
                        LOGOUT
                    </button>
                </div>
            </header>

            {/* Main content - responsive */}
            <div className='flex-1 flex overflow-hidden'>
                {/* Desktop: 3-column layout */}
                {/* Mobile: Show either list OR content */}

                {/* Desktop sidebar - hidden on mobile */}
                <aside className='hidden md:flex w-14 border-r border-border flex-col shrink-0'>
                    <div className='flex-1 overflow-y-auto py-2'>
                        <div className='flex flex-col items-center gap-1 pb-2 border-b border-border mx-2 mb-2'>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <div className='w-10 h-10 border border-border rounded-md flex items-center justify-center cursor-pointer hover:bg-secondary transition-colors'>
                                        <div className='grid grid-cols-3 gap-0.5'>
                                            {[1, 0, 1, 0, 1, 0, 1, 0, 1].map(
                                                (filled, i) => (
                                                    <div
                                                        key={i}
                                                        className={`w-1.5 h-1.5 ${
                                                            filled
                                                                ? 'bg-foreground'
                                                                : 'bg-transparent'
                                                        }`}
                                                    />
                                                ),
                                            )}
                                        </div>
                                    </div>
                                </TooltipTrigger>
                                <TooltipContent side='right'>
                                    Current: {currentAccount.address}
                                </TooltipContent>
                            </Tooltip>
                        </div>
                        <div className='flex flex-col items-center gap-2 mt-2'>
                            {accounts.map((account, index) => (
                                <Tooltip key={account.address}>
                                    <TooltipTrigger asChild>
                                        <div
                                            onClick={() =>
                                                handleSelectAccount(index)
                                            }
                                            className={`w-8 h-8 border rounded flex items-center justify-center cursor-pointer transition-colors text-xs font-mono ${
                                                selectedAccount === index
                                                    ? 'border-primary bg-primary/10 text-primary'
                                                    : 'border-border text-muted-foreground hover:bg-secondary'
                                            }`}
                                        >
                                            {index + 1}
                                        </div>
                                    </TooltipTrigger>
                                    <TooltipContent side='right'>
                                        {account.address}
                                    </TooltipContent>
                                </Tooltip>
                            ))}
                        </div>
                    </div>
                    <div className='p-2 border-t border-border flex justify-center'>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <button className='w-8 h-8 border border-dashed border-border rounded flex items-center justify-center hover:bg-secondary hover:border-solid transition-colors'>
                                    <Plus className='w-4 h-4 text-muted-foreground' />
                                </button>
                            </TooltipTrigger>
                            <TooltipContent side='right'>
                                Derive new address
                            </TooltipContent>
                        </Tooltip>
                    </div>
                </aside>

                {/* Email list - full width on mobile when viewing list */}
                <div
                    className={`${
                        mobileView === 'list' ? 'flex' : 'hidden'
                    } md:flex w-full md:w-80 border-r border-border flex-col shrink-0`}
                >
                    <div className='px-4 py-3 border-b border-border flex items-center justify-between'>
                        <span className='tech-label'>
                            Inbox ({emails.length})
                        </span>
                    </div>
                    <div className='flex-1 overflow-y-auto'>
                        {emails.length === 0 ? (
                            <div className='px-4 py-12 text-center'>
                                <div className='radar-container mx-auto'>
                                    <div className='radar-sweep' />
                                    <div className='radar-sweep' />
                                    <div className='radar-sweep' />
                                    <div className='radar-dot' />
                                </div>
                                <span className='font-mono text-xs text-muted-foreground'>
                                    [ NO_INBOUND_DATA ]
                                </span>
                            </div>
                        ) : (
                            emails.map((email) => (
                                <div
                                    key={email.id}
                                    onClick={() => handleSelectEmail(email)}
                                    className={`email-row group ${
                                        selectedEmail?.id === email.id
                                            ? 'email-row-active'
                                            : ''
                                    }`}
                                >
                                    <div className='flex items-start justify-between gap-2'>
                                        <div className='text-sm font-medium truncate mb-1 flex-1'>
                                            {email.from}
                                        </div>
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleDeleteEmail(email.id);
                                            }}
                                            className='p-1 opacity-0 group-hover:opacity-100 hover:bg-destructive/10 hover:text-destructive rounded transition-all shrink-0'
                                        >
                                            <Trash2 className='w-3.5 h-3.5' />
                                        </button>
                                    </div>
                                    <div className='text-sm truncate text-muted-foreground mb-2'>
                                        {email.subject}
                                    </div>
                                    <div className='font-mono text-xs text-muted-foreground'>
                                        TTL: {email.ttl}
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                </div>

                {/* Email content - full width on mobile when viewing content */}
                <main
                    className={`${
                        mobileView === 'content' ? 'flex' : 'hidden'
                    } md:flex flex-1 flex-col overflow-hidden`}
                >
                    <div className='px-4 py-3 border-b border-border flex items-center justify-between'>
                        {/* Mobile back button */}
                        <button
                            onClick={() => setMobileView('list')}
                            className='md:hidden p-1 -ml-1 mr-2 text-muted-foreground hover:text-foreground'
                        >
                            <ArrowLeft className='w-5 h-5' />
                        </button>
                        <span className='tech-label'>Message</span>
                        {selectedEmail && (
                            <button
                                onClick={() =>
                                    handleDeleteEmail(selectedEmail.id)
                                }
                                className='p-1 hover:bg-destructive/10 hover:text-destructive rounded transition-colors'
                            >
                                <Trash2 className='w-4 h-4 text-muted-foreground' />
                            </button>
                        )}
                    </div>
                    <div className='flex-1 overflow-y-auto'>
                        {selectedEmail ? (
                            <div className='p-4 md:p-6'>
                                <div className='border-b border-border pb-4 mb-4 md:mb-6'>
                                    <h2 className='text-lg md:text-xl font-semibold mb-3 md:mb-4'>
                                        {selectedEmail.subject}
                                    </h2>
                                    <div className='font-mono text-xs md:text-sm space-y-1'>
                                        <p>
                                            <span className='text-muted-foreground'>
                                                From:{' '}
                                            </span>
                                            <span className='break-all'>
                                                {selectedEmail.from}
                                            </span>
                                        </p>
                                        <p>
                                            <span className='text-muted-foreground'>
                                                Time:{' '}
                                            </span>
                                            <span className='break-all'>
                                                {selectedEmail.timestamp.toISOString()}
                                            </span>
                                        </p>
                                        <p>
                                            <span className='text-muted-foreground'>
                                                TTL:{' '}
                                            </span>
                                            <span className='text-primary'>
                                                {selectedEmail.ttl}
                                            </span>
                                        </p>
                                    </div>
                                </div>
                                <div className='whitespace-pre-wrap leading-relaxed text-sm md:text-base'>
                                    {selectedEmail.body}
                                </div>
                            </div>
                        ) : (
                            <div className='flex-1 flex items-center justify-center h-full p-4'>
                                <div className='text-center'>
                                    <div className='radar-container mx-auto'>
                                        <div className='radar-sweep' />
                                        <div className='radar-sweep' />
                                        <div className='radar-sweep' />
                                        <div className='radar-dot' />
                                    </div>
                                    <p className='font-mono text-sm md:text-lg text-muted-foreground mb-2'>
                                        [ STATUS: AWAITING_INBOUND_DATA ]
                                    </p>
                                    <p className='font-mono text-xs md:text-sm text-muted-foreground'>
                                        // Select an email to view
                                    </p>
                                </div>
                            </div>
                        )}
                    </div>
                </main>
            </div>
        </div>
    );
};

export default Inbox;
