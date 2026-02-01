import { SunIcon, MoonIcon } from "@phosphor-icons/react";
import { useTheme } from "@/components/theme-provider";

export function ModeToggle() {
    const { setTheme } = useTheme();

    return (
        <div className="relative flex items-center h-5">
            <button
                type="button"
                onClick={() => {
                    const newTheme =
                        localStorage.getItem("coresend-theme") === "dark"
                            ? "light"
                            : "dark";
                    setTheme(newTheme);
                }}
                aria-label="Toggle theme"
                className="flex items-center justify-center"
            >
                <SunIcon
                    weight="duotone"
                    className="h-4 w-4 text-muted-foreground hover:text-foreground dark:hidden transition-colors"
                />
                <MoonIcon
                    weight="duotone"
                    className="h-4 w-4 text-muted-foreground hover:text-foreground hidden dark:block transition-colors"
                />
            </button>
        </div>
    );
}
