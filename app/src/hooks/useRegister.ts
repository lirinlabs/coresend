import { useRegisterAddress } from '@/api/generated';

export const useRegister = () =>
    useRegisterAddress({
        mutation: {
            onSuccess: (response) => {
                if (response.status === 200) {
                    console.log('Address registered successfully');
                } else {
                    console.error('Failed to register address:', response);
                }
            },
        },
    });
