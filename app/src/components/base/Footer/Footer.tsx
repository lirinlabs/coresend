import Typography from '@/components/base/Typography/typography';

export const Footer = () => {
    return (
        <footer className='border-t border-border px-6 py-4'>
            <div className='max-w-7xl mx-auto flex flex-wrap items-center justify-between gap-4'>
                <div className='flex items-center gap-6'>
                    <div className='flex items-center gap-2'>
                        <span className='w-2 h-2 rounded-full bg-primary animate-pulse'></span>
                        <Typography
                            font='mono'
                            text='xs'
                            color='muted'
                            as='span'
                        >
                            Status:{' '}
                            <Typography text='xs' color='foreground' as='span'>
                                REDIS_CONNECTED
                            </Typography>
                        </Typography>
                    </div>
                    <Typography font='mono' text='xs' color='muted' as='span'>
                        Uptime:{' '}
                        <Typography text='xs' color='foreground' as='span'>
                            99.9%
                        </Typography>
                    </Typography>
                    <Typography font='mono' text='xs' color='muted' as='span'>
                        Block Height:{' '}
                        <Typography text='xs' color='foreground' as='span'>
                            [N/A]
                        </Typography>
                    </Typography>
                </div>
                <Typography font='mono' text='xs' color='muted' as='span'>
                    v0.1.0-alpha
                </Typography>
            </div>
        </footer>
    );
};
