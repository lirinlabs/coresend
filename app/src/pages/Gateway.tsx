import Button from '@/components/base/Button/Button';
import { GatewayHeader } from '@/components/base/Header/GatewayHeader';
import { SeedBox } from '@/components/base/Seeds/SeedBox';
import Typography from '@/components/base/Typography/typography';
import { useNavigate } from 'react-router-dom';

const Gateway = () => {
    const navigate = useNavigate();

    return (
        <div className='w-full h-dvh flex flex-col'>
            <GatewayHeader />

            <div className='max-w-7xl mx-auto w-full px-4 flex-1 flex flex-col justify-center items-center'>
                <div className='flex items-center flex-col mb-8'>
                    <Typography
                        weight='bold'
                        text='3xl'
                        color='foreground'
                        align='center'
                        className='text-4xl mb-4'
                        as='h1'
                    >
                        Authenticate Session
                    </Typography>
                    <Typography
                        text='sm'
                        font='mono'
                        color='muted'
                        align='center'
                    >
                        Enter your 12-word seed phrase to derive inbox address.
                    </Typography>
                </div>
                {/* Seeds */}
                <SeedBox
                    seedWords={new Array(12).fill('')}
                    onChangeWord={() => {}}
                    onKeyDownWord={() => {}}
                />
                {/* Actions */}
                <div className='flex flex-col md:flex-row gap-4 w-full'>
                    <Button
                        variant='primary'
                        size='md'
                        className='flex-1 w-full'
                        onClick={() => {
                            navigate('/inbox');
                        }}
                    >
                        Unlock Inbox
                    </Button>
                    <Button
                        variant='secondary'
                        size='md'
                        className='flex-1 w-full'
                        // TODO: onClick = call api to generate new seed phrase
                    >
                        Generate New Seed
                    </Button>
                </div>
                <Typography
                    weight='light'
                    text='xs'
                    tracking='tight'
                    font='mono'
                    color='muted'
                    className='mt-12'
                    align='center'
                >
                    [ NOTICE: Seed phrase is processed client-side only. Never
                    transmitted. ]
                </Typography>
            </div>
        </div>
    );
};

export default Gateway;
