import { useHealth } from '@/hooks/useHealth';

export function DemoHealth() {
    const { data, isPending, isError, error } = useHealth();

    if (isPending)
        return (
            <p className='text-xs text-muted-foreground'>
                Checking API health...
            </p>
        );

    if (isError)
        return (
            <p className='text-xs text-destructive'>
                Error:{' '}
                {error instanceof Error ? error.message : 'Unknown error'}
            </p>
        );

    return (
        <div className='w-full max-w-xl rounded-lg border border-border bg-card p-4'>
            <h3 className='mb-2 text-sm font-semibold'>API health</h3>
            <p className='text-xs text-muted-foreground'>
                Status: {data?.status ?? 'unknown'}
            </p>
            <ul className='mt-2 list-disc pl-5 text-xs text-muted-foreground'>
                {data?.services &&
                    Object.entries(data.services).map(([name, status]) => (
                        <li key={name}>
                            {name}: {status}
                        </li>
                    ))}
            </ul>
        </div>
    );
}
