import { cva, type VariantProps } from "class-variance-authority";
import { forwardRef, type ReactNode, type ButtonHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
    [
        "inline-flex items-center justify-center gap-2",
        "font-semibold uppercase tracking-wide",
        "border border-foreground",
        "transition-all duration-150 ease-out",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        "disabled:pointer-events-none disabled:opacity-50",
    ],
    {
        variants: {
            variant: {
                primary: [
                    "bg-primary text-primary-foreground",
                    "shadow-hard",
                    "hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-hard-sm",
                    "active:translate-x-[4px] active:translate-y-[4px] active:shadow-none",
                ],
                secondary: [
                    "bg-background text-foreground",
                    "shadow-hard",
                    "hover:translate-x-[2px] hover:translate-y-[2px] hover:bg-secondary hover:shadow-hard-sm",
                    "active:translate-x-[4px] active:translate-y-[4px] active:shadow-none",
                ],
                outline: [
                    "bg-transparent text-foreground",
                    "shadow-hard",
                    "hover:translate-x-[2px] hover:translate-y-[2px] hover:bg-accent hover:shadow-hard-sm",
                    "active:translate-x-[4px] active:translate-y-[4px] active:shadow-none",
                ],
                ghost: [
                    "bg-transparent text-foreground",
                    "border-transparent shadow-none",
                    "hover:text-accent",
                ],
            },
            size: {
                sm: "px-4 py-2 text-xs",
                md: "px-6 py-3 text-sm",
                lg: "px-8 py-4 text-base",
            },
        },
        defaultVariants: {
            variant: "primary",
            size: "md",
        },
    }
);

const LoadingSpinner = ({ className }: { className?: string }) => (
    <svg
        className={cn("animate-spin", className)}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        aria-hidden="true"
    >
        <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
        />
        <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
    </svg>
);

export interface ButtonProps
    extends ButtonHTMLAttributes<HTMLButtonElement>,
        VariantProps<typeof buttonVariants> {
    /** Shows a loading spinner and disables the button */
    loading?: boolean;
    /** Icon to display on the left side of the button text */
    leftIcon?: ReactNode;
    /** Icon to display on the right side of the button text */
    rightIcon?: ReactNode;
    /** Content to render inside the button */
    children: ReactNode;
    /** Additional class names to apply */
    className?: string;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
    (
        {
            variant = "primary",
            size = "md",
            loading = false,
            disabled = false,
            leftIcon,
            rightIcon,
            children,
            className,
            ...props
        },
        ref
    ) => {
        const isDisabled = disabled || loading;

        // Determine icon size based on button size
        const iconSize = {
            sm: "h-3 w-3",
            md: "h-4 w-4",
            lg: "h-5 w-5",
        }[size ?? "md"];

        return (
            <button
                ref={ref}
                className={cn(buttonVariants({ variant, size }), className)}
                disabled={isDisabled}
                aria-busy={loading}
                {...props}
            >
                {loading ? (
                    <LoadingSpinner className={iconSize} />
                ) : (
                    leftIcon && (
                        <span className={cn("shrink-0", iconSize)}>
                            {leftIcon}
                        </span>
                    )
                )}
                {children}
                {!loading && rightIcon && (
                    <span className={cn("shrink-0", iconSize)}>
                        {rightIcon}
                    </span>
                )}
            </button>
        );
    }
);

Button.displayName = "Button";

export default Button;
