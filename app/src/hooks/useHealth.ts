import { useHealthCheck } from '@/api/generated';

export const useHealth = () =>
    useHealthCheck({
        query: {
            select: (response) => response.data,
            staleTime: 30_000,
            refetchInterval: 30_000,
        },
    });
