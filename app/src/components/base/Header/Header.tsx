import { Logo } from "../Logo/Logo";
import { ModeToggle } from "../ModeToggle/ModeToggle";

export const Header = () => {
    return (
        <header className="w-full p-4">
            <div className="max-w-7xl mx-auto flex items-center justify-between">
                <Logo navigate />
                <ModeToggle />
            </div>
        </header>
    );
};
