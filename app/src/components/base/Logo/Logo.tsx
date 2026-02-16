import { useNavigate } from 'react-router-dom';
import Typography from '../Typography/typography';

interface LogoProps {
    navigate?: boolean;
}

export const Logo = ({ navigate }: LogoProps) => {
    const nav = useNavigate();

    const logoContent = (
        <Typography
            weight='semibold'
            text='sm'
            tracking='tight'
            font='mono'
            color='foreground'
            as='span'
            transform='uppercase'
            className='leading-none'
        >
            CORESEND
        </Typography>
    );

    if (navigate) {
        return (
            <button
                type='button'
                onClick={() => nav('/')}
                className='cursor-pointer'
            >
                {logoContent}
            </button>
        );
    }

    return logoContent;
};
