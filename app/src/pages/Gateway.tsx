import { useEffect, useState } from 'react';
import Button from '@/components/base/Button/Button';
import { GatewayHeader } from '@/components/base/Header/GatewayHeader';
import { SeedBox } from '@/components/base/Seeds/SeedBox';
import Typography from '@/components/base/Typography/typography';
import { deriveIdentityFromMnemonic } from '@/lib/crypto/deriveIdentityFromMnemonic';
import {
    validateSeedWords,
    generateMnemonicWords,
} from '@/lib/crypto/validateMnemonic';
import { useNavigate } from 'react-router-dom';
import { useRegister } from '@/hooks/useRegister';
import { useIdentityStore } from '@/lib/stores/identityStore';

const WORD_COUNT = 12;
const ENTROPY_BITS = 128;

const Gateway = () => {
    const navigate = useNavigate();
    const { mutateAsync: register, isPending } = useRegister();
    const [seedWords, setSeedWords] = useState<string[]>(
        Array(WORD_COUNT).fill(''),
    );
    const [error, setError] = useState<string | null>(null);
    const [invalidIndices, setInvalidIndices] = useState<number[]>([]);
    const [showConfirmDialog, setShowConfirmDialog] = useState(false);

    const updateWordAtIndex = (index: number, value: string) => {
        setSeedWords((prev) => {
            const next = [...prev];
            next[index] = value;
            return next;
        });
        if (error) setError(null);
        if (invalidIndices.length > 0) setInvalidIndices([]);
    };

    const handleUnlockInbox = async () => {
        setError(null);
        setInvalidIndices([]);

        try {
            const validation = validateSeedWords(seedWords);

            if (!validation.valid) {
                setError(validation.error!);
                setInvalidIndices(validation.invalidIndices || []);
                return;
            }

            const mnemonic = seedWords.join(' ').trim();

            const identity = deriveIdentityFromMnemonic(mnemonic);

            useIdentityStore.getState().setIdentity(identity);

            await register({ address: identity.address });

            sessionStorage.setItem(
                'identity',
                JSON.stringify({
                    address: identity.address,
                    publicKey: Array.from(identity.publicKey),
                }),
            );

            navigate('/inbox');
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to register');
        }
    };

    const handleGenerateNewSeed = () => {
        const hasWords = seedWords.some((word) => word.trim() !== '');

        if (hasWords) {
            setShowConfirmDialog(true);
        } else {
            generateNewSeed();
        }
    };

    const generateNewSeed = () => {
        const words = generateMnemonicWords(ENTROPY_BITS);
        setSeedWords(
            Array.from(
                { length: WORD_COUNT },
                (_, index) => words[index] ?? '',
            ),
        );
        setError(null);
        setInvalidIndices([]);
        setShowConfirmDialog(false);
    };

    const isButtonDisabled =
        isPending || seedWords.some((word) => !word.trim());

    useEffect(() => {
        return () => setSeedWords(Array(WORD_COUNT).fill(''));
    }, []);

    return (
        <div className='w-full h-dvh flex flex-col'>
            <GatewayHeader />

            <div className='max-w-7xl mx-auto w-full px-4 flex-1 flex flex-col justify-center items-center'>
                <div className='flex items-center flex-col mb-8'>
                    <Typography
                        weight='bold'
                        text='3xl'
                        color='foreground'
                        align='center'
                        className='text-4xl mb-4'
                        as='h1'
                    >
                        Authenticate Session
                    </Typography>
                    <Typography
                        text='sm'
                        font='mono'
                        color='muted'
                        align='center'
                    >
                        Enter your 12-word seed phrase to derive inbox address.
                    </Typography>
                </div>
                <div className='flex flex-col mb-6'>
                    <SeedBox
                        seedWords={seedWords}
                        onChangeWord={updateWordAtIndex}
                        onKeyDownWord={(
                            _index: number,
                            event: React.KeyboardEvent<HTMLInputElement>,
                        ) => {
                            if (event.key === 'Enter') {
                                handleUnlockInbox();
                            }
                        }}
                        invalidIndices={invalidIndices}
                    />
                    {error && (
                        <div className='w-full mt-4'>
                            <Typography
                                text='sm'
                                color='destructive'
                                align='center'
                                font='mono'
                            >
                                ⚠ {error}
                            </Typography>
                        </div>
                    )}
                </div>

                <div className='flex flex-col md:flex-row gap-4 w-full'>
                    <Button
                        variant='primary'
                        size='md'
                        className='flex-1 w-full'
                        onClick={handleUnlockInbox}
                        disabled={isButtonDisabled}
                        style={{ opacity: isButtonDisabled ? 0.5 : 1 }}
                    >
                        {isPending ? 'Unlocking...' : 'Unlock Inbox'}
                    </Button>
                    <Button
                        variant='secondary'
                        size='md'
                        className='flex-1 w-full'
                        onClick={handleGenerateNewSeed}
                        disabled={isPending}
                        style={{ opacity: isPending ? 0.5 : 1 }}
                    >
                        Generate New Seed
                    </Button>
                </div>
                <Typography
                    weight='light'
                    text='xs'
                    tracking='tight'
                    font='mono'
                    color='muted'
                    className='mt-12'
                    align='center'
                >
                    [ NOTICE: Seed phrase is processed client-side only. Never
                    transmitted. ]
                </Typography>
            </div>

            {showConfirmDialog && (
                <div
                    className='fixed inset-0 bg-black/50 flex items-center justify-center z-50'
                    onClick={() => setShowConfirmDialog(false)}
                    onKeyDown={(e) => {
                        if (e.key === 'Escape') setShowConfirmDialog(false);
                    }}
                    role='dialog'
                    aria-modal='true'
                    aria-labelledby='dialog-title'
                >
                    <div
                        className='bg-background border border-border rounded-lg p-6 max-w-md mx-4'
                        onClick={(e) => e.stopPropagation()}
                        onKeyDown={(e) => e.stopPropagation()}
                        role='document'
                    >
                        <Typography
                            text='lg'
                            weight='bold'
                            color='foreground'
                            className='mb-4'
                            as='h2'
                            id='dialog-title'
                        >
                            Generate New Seed?
                        </Typography>
                        <Typography text='sm' color='muted' className='mb-6'>
                            This will replace your current seed phrase. Make
                            sure you've saved it if needed.
                        </Typography>
                        <div className='flex gap-3'>
                            <Button
                                variant='secondary'
                                size='sm'
                                className='flex-1'
                                onClick={() => setShowConfirmDialog(false)}
                            >
                                Cancel
                            </Button>
                            <Button
                                variant='primary'
                                size='sm'
                                className='flex-1'
                                onClick={generateNewSeed}
                            >
                                Generate
                            </Button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Gateway;
