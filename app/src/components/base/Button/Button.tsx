import { cva, type VariantProps } from "class-variance-authority";
import { forwardRef, type ReactNode, type ButtonHTMLAttributes } from "react";
import { cn } from "@/lib/utils";
import Typography from "@/components/base/Typography/typography";

const buttonVariants = cva(
    // Base styles for all buttons
    [
        "inline-flex items-center justify-center gap-2",
        "border border-foreground",
        "transition-all duration-150 ease-out",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        "disabled:pointer-events-none disabled:opacity-50",
    ],
    {
        variants: {
            variant: {
                primary: [
                    "bg-primary text-white",
                    "shadow-[4px_4px_0px_0px_#000000]",
                    "hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_#000000]",
                    "active:translate-x-[4px] active:translate-y-[4px] active:shadow-none",
                ],
                secondary: [
                    "bg-background text-foreground",
                    "shadow-[4px_4px_0px_0px_#000000]",
                    "hover:translate-x-[2px] hover:translate-y-[2px] hover:bg-secondary hover:shadow-[2px_2px_0px_0px_#000000]",
                    "active:translate-x-[4px] active:translate-y-[4px] active:shadow-none",
                ],
                outline: [
                    "bg-transparent text-foreground",
                    "shadow-[4px_4px_0px_0px_#000000]",
                    "hover:translate-x-[2px] hover:translate-y-[2px] hover:bg-accent hover:shadow-[2px_2px_0px_0px_#000000]",
                    "active:translate-x-[4px] active:translate-y-[4px] active:shadow-none",
                ],
                ghost: [
                    "bg-transparent text-foreground",
                    "border-transparent shadow-none",
                    "hover:bg-accent hover:text-accent-foreground",
                    "active:bg-accent/80",
                ],
            },
            size: {
                sm: "px-4 py-2",
                md: "px-6 py-3",
                lg: "px-8 py-4",
            },
        },
        defaultVariants: {
            variant: "primary",
            size: "md",
        },
    }
);

// Map button size to typography text size
const sizeToTextSize = {
    sm: "xs",
    md: "sm",
    lg: "base",
} as const;

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
                <Typography
                    as="span"
                    text={sizeToTextSize[size ?? "md"]}
                    weight="semibold"
                    transform="uppercase"
                    tracking="wide"
                    color={variant === "primary" ? "primary-foreground" : "foreground"}
                >
                    {children}
                </Typography>
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
