import { useTheme } from '@/components/theme-provider';
import BrightnessDownIcon from '@/components/ui/brightness-down-icon';
import MoonIcon from '@/components/ui/moon-icon';

export function ModeToggle() {
    const { setTheme } = useTheme();

    return (
        <div className='relative flex items-center h-5'>
            <button
                type='button'
                onClick={() => {
                    const newTheme =
                        localStorage.getItem('coresend-theme') === 'dark'
                            ? 'light'
                            : 'dark';
                    setTheme(newTheme);
                }}
                aria-label='Toggle theme'
                className='flex items-center justify-center cursor-pointer'
            >
                <BrightnessDownIcon className='h-4 w-4 text-muted-foreground hover:text-primary dark:hidden transition-colors' />
                <MoonIcon className='h-4 w-4 text-muted-foreground hover:text-primary hidden dark:block transition-colors' />
            </button>
        </div>
    );
}
