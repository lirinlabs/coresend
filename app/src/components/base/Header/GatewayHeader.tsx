import { useNavigate } from 'react-router-dom';
import { ModeToggle } from '../ModeToggle/ModeToggle';
import Typography from '../Typography/typography';

export const GatewayHeader = () => {
    const navigate = useNavigate();
    return (
        <header className='w-full p-4 border-b border-border'>
            <div className='max-w-7xl mx-auto flex items-center justify-between'>
                <button
                    type='button'
                    onClick={() => navigate('/')}
                    className='cursor-pointer'
                >
                    <Typography
                        weight='semibold'
                        text='sm'
                        tracking='tight'
                        font='mono'
                        color='foreground'
                        as='span'
                        transform='uppercase'
                        className='leading-none hover:text-primary cursor-pointer'
                    >
                        ← RETURN_TO_ORIGIN
                    </Typography>
                </button>
                <ModeToggle />
            </div>
        </header>
    );
};
