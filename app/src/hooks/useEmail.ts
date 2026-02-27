import { useGetEmail } from '@/api/generated';

export const useEmail = () => {
    useGetEmail('email@email.com', 'email', {
        query: {
            enabled: false,
            staleTime: 5000,
        },
    });

    return useGetEmail;
};
