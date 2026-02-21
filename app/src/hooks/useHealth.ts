import { type ApiHealthResponse, useGetApiHealth } from '@/api/generated';

export const useHealth = () =>
    useGetApiHealth<ApiHealthResponse>({
        query: {
            select: (response) => response.data,
            staleTime: 30_000,
            refetchInterval: 30_000,
        },
    });
