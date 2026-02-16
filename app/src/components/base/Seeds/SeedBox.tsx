import SeedSlot from './SeedSlot';

export const SeedBox = ({
    seedWords,
    onChangeWord,
    onKeyDownWord,
}: {
    seedWords: string[];
    onChangeWord: (index: number, value: string) => void;
    onKeyDownWord: (
        index: number,
        e: React.KeyboardEvent<HTMLInputElement>,
    ) => void;
}) => {
    return (
        <div className='border border-foreground p-6 mb-8 bg-background shadow-hard'>
            <div className='grid grid-cols-2 md:grid-cols-3 gap-3'>
                {seedWords.map((word, index) => (
                    <SeedSlot
                        key={index}
                        index={index}
                        value={word}
                        onChange={(value) => onChangeWord(index, value)}
                        onKeyDown={(e) => onKeyDownWord(index, e)}
                    />
                ))}
            </div>
        </div>
    );
};
