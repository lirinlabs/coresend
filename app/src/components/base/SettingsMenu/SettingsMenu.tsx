import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ModeToggle } from '../ModeToggle/ModeToggle';
import GearIcon from '@/components/ui/gear-icon';
import { copyToClipboard } from '@/lib/utils';
import { bitcoinAddress, evmAddress, solanaAddress } from '@/lib/consts';
import GithubIcon from '@/components/ui/github-icon';
import ExternalLinkIcon from '@/components/ui/external-link-icon';
import HeartIcon from '@/components/ui/heart-icon';
import WalletIcon from '@/components/ui/wallet-icon';
import Typography from '../Typography/typography';

export const SettingsMenu = () => {
    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <button
                    className='text-muted-foreground hover:text-foreground transition-colors'
                    aria-label='Settings'
                >
                    <GearIcon className='text-muted-foreground hover:text-primary transition-colors h-4 w-4' />
                </button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align='end' className='w-64'>
                <DropdownMenuLabel>
                    <Typography
                        text='sm'
                        weight='bold'
                        color='foreground'
                        tracking='tight'
                        font='mono'
                    >
                        Settings
                    </Typography>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />

                <div className='px-2 py-1.5 flex items-center justify-between'>
                    <Typography
                        text='sm'
                        weight='bold'
                        color='foreground'
                        tracking='tight'
                    >
                        Theme
                    </Typography>
                    <ModeToggle />
                </div>

                <DropdownMenuSeparator />
                <DropdownMenuLabel>
                    <Typography
                        text='sm'
                        weight='bold'
                        color='foreground'
                        tracking='tight'
                    >
                        Links
                    </Typography>
                </DropdownMenuLabel>

                <DropdownMenuItem asChild>
                    <a
                        href='https://github.com/coresend'
                        target='_blank'
                        rel='noopener noreferrer'
                        className='flex items-center gap-2 cursor-pointer'
                    >
                        <GithubIcon className='w-4 h-4' />
                        <Typography
                            text='sm'
                            weight='bold'
                            color='foreground'
                            tracking='tight'
                        >
                            Github
                        </Typography>
                        <ExternalLinkIcon className='w-3 h-3 ml-auto' />
                    </a>
                </DropdownMenuItem>

                <DropdownMenuSeparator />
                <DropdownMenuLabel className='font-mono text-xs flex items-center gap-1'>
                    <HeartIcon className='w-4 h-4 text-primary' />
                    <Typography
                        text='sm'
                        weight='bold'
                        color='foreground'
                        tracking='tight'
                    >
                        Support / Tips
                    </Typography>
                </DropdownMenuLabel>

                <DropdownMenuItem
                    onClick={() => copyToClipboard(solanaAddress)}
                    className='flex flex-col items-start gap-0.5 cursor-pointer'
                >
                    <div className='flex items-center gap-1'>
                        <WalletIcon className='w-4 h-4' />
                        <Typography
                            text='sm'
                            weight='bold'
                            color='foreground'
                            tracking='tight'
                        >
                            Solana
                        </Typography>
                    </div>
                    <Typography
                        text='xs'
                        truncate
                        tracking='tight'
                        weight='medium'
                        color='muted'
                        font='mono'
                        className='w-full'
                    >
                        {solanaAddress}
                    </Typography>
                </DropdownMenuItem>

                <DropdownMenuItem
                    onClick={() => copyToClipboard(evmAddress)}
                    className='flex flex-col items-start gap-0.5 cursor-pointer'
                >
                    <div className='flex items-center gap-1'>
                        <WalletIcon className='w-4 h-4' />
                        <Typography
                            text='sm'
                            weight='bold'
                            color='foreground'
                            tracking='tight'
                        >
                            EVM (ETH/Base/etc)
                        </Typography>
                    </div>

                    <Typography
                        text='xs'
                        truncate
                        tracking='tight'
                        weight='medium'
                        color='muted'
                        font='mono'
                        className='w-full'
                    >
                        {evmAddress}
                    </Typography>
                </DropdownMenuItem>

                <DropdownMenuItem
                    onClick={() => copyToClipboard(evmAddress)}
                    className='flex flex-col items-start gap-0.5 cursor-pointer'
                >
                    <div className='flex items-center gap-1'>
                        <WalletIcon className='w-4 h-4' />
                        <Typography
                            text='sm'
                            weight='bold'
                            color='foreground'
                            tracking='tight'
                        >
                            Bitcoin
                        </Typography>
                    </div>

                    <Typography
                        text='xs'
                        truncate
                        tracking='tight'
                        weight='medium'
                        color='muted'
                        font='mono'
                        className='w-full'
                    >
                        {bitcoinAddress}
                    </Typography>
                </DropdownMenuItem>

                <DropdownMenuSeparator />
                <div className='px-2 py-2 text-center'>
                    <Typography
                        text='xs'
                        font='mono'
                        tracking='tight'
                        color='muted'
                        align='center'
                    >
                        CORESEND v0.1.0-alpha
                    </Typography>
                    <Typography
                        text='xs'
                        font='mono'
                        tracking='tight'
                        color='muted'
                        align='center'
                        className='mt-1'
                    >
                        Built with BIP39 deterministic logic
                    </Typography>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
};
