import { useGetApiHealth } from '@/api/generated';

export function DemoHealth() {
    const { data, isLoading, error } = useGetApiHealth();

    if (isLoading) return <p>Loading...</p>;
    if (error)
        return (
            <p>
                Error:{' '}
                {error instanceof Error ? error.message : 'Unknown error'}
            </p>
        );

    const healthData = data?.data;

    return (
        <div className='p-4 border rounded-lg'>
            <h3 className='font-semibold mb-2'>API Health</h3>
            <p>Status: {healthData?.status}</p>
            <p>Services:</p>
            <ul className='list-disc pl-5'>
                {healthData?.services &&
                    Object.entries(healthData.services).map(
                        ([name, status]) => (
                            <li key={name}>
                                {name}: {status}
                            </li>
                        ),
                    )}
            </ul>
        </div>
    );
}
