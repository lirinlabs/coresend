import Typography from '../Typography/typography';
import Spinner from '../../../assets/Loader-F50.svg';
export const MessagePanelEmpty = () => {
    return (
        <div className='flex-1 flex flex-col gap-2.5 items-center justify-center p-4'>
            <div className='mx-auto mb-8'>
                <img src={Spinner} alt='Spinner' className='w-14' />
            </div>
            <Typography
                text='base'
                font='mono'
                color='accent-foreground'
                tracking='wide'
            >
                [ STATUS: AWAITING_INBOUND_DATA ]
            </Typography>
            <Typography text='xs' font='mono' color='muted'>
                // No packet selected.
            </Typography>
        </div>
    );
};
