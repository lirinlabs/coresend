import { Logo } from "../Logo/Logo";
import { ModeToggle } from "../ModeToggle/ModeToggle";

export const InboxHeader = () => {
    return (
        <header className="w-full py-2">
            <div className="max-w-7xl mx-auto flex items-center justify-between">
                <Logo />
                <ModeToggle />
            </div>
        </header>
    );
};
