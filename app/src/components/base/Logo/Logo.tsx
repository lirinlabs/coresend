import { useNavigate } from 'react-router-dom';
import Typography from '../Typography/typography';
import { useTheme } from '@/components/theme-provider';

interface LogoProps {
    navigate?: boolean;
}

export const Logo = ({ navigate }: LogoProps) => {
    const nav = useNavigate();
    const { theme } = useTheme();

    const logoSrc =
        theme === 'dark'
            ? '/src/assets/Logo-FFF.svg'
            : '/src/assets/Logo-F50.svg';

    const logoContent = (
        <div className='flex items-center gap-3'>
            <img src={logoSrc} alt='Coresend' className='h-6 w-auto' />
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
        </div>
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
