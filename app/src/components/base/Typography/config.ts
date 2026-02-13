import { cva } from "class-variance-authority";

export const typography = cva("antialiased", {
    variants: {
        text: {
            xs: "text-xs leading-tight",
            sm: "text-sm leading-snug",
            base: "text-base leading-normal",
            lg: "text-lg leading-relaxed",
            xl: "text-xl leading-relaxed",
            "2xl": "text-2xl leading-tight",
            "3xl": "text-3xl leading-tight",
            "4xl": "text-4xl leading-tight",
            "5xl": "text-5xl leading-none",
            "6xl": "text-6xl leading-none",
            "7xl": "text-7xl leading-none",
            "8xl": "text-8xl leading-none",
            "9xl": "text-9xl leading-none",
        },

        textLg: {
            xs: "lg:text-xs leading-tight",
            sm: "lg:text-sm leading-snug",
            base: "lg:text-base leading-normal",
            lg: "lg:text-lg leading-relaxed",
            xl: "lg:text-xl leading-relaxed",
            "2xl": "lg:text-2xl leading-tight",
            "3xl": "lg:text-3xl leading-tight",
            "4xl": "lg:text-4xl leading-tight",
            "5xl": "lg:text-5xl leading-none",
            "6xl": "lg:text-6xl leading-none",
            "7xl": "lg:text-7xl leading-none",
            "8xl": "lg:text-8xl leading-none",
            "9xl": "lg:text-9xl leading-none",
        },

        color: {
            foreground: "text-foreground",
            muted: "text-muted-foreground",
            primary: "text-primary",
            secondary: "text-secondary",
            accent: "text-accent",
            destructive: "text-destructive",
            "primary-foreground": "text-primary-foreground",
            "secondary-foreground": "text-secondary-foreground",
            "accent-foreground": "text-accent-foreground",
        },

        font: {
            sans: "font-sans",
            mono: "font-mono",
        },

        weight: {
            thin: "font-thin",
            extralight: "font-extralight",
            light: "font-light",
            normal: "font-normal",
            medium: "font-medium",
            semibold: "font-semibold",
            bold: "font-bold",
            extrabold: "font-extrabold",
            black: "font-black",
        },

        align: {
            left: "text-left",
            center: "text-center",
            right: "text-right",
            justify: "text-justify",
        },

        transform: {
            uppercase: "uppercase",
            lowercase: "lowercase",
            capitalize: "capitalize",
            "normal-case": "normal-case",
        },

        tracking: {
            tighter: "tracking-tighter",
            tight: "tracking-tight",
            normal: "tracking-normal",
            wide: "tracking-wide",
            wider: "tracking-wider",
            widest: "tracking-widest",
        },

        select: {
            false: "select-none",
            true: "select-text",
        },

        nowrap: {
            false: "whitespace-normal",
            true: "whitespace-nowrap",
        },

        ellipsis: {
            true: "text-ellipsis overflow-hidden whitespace-nowrap",
        },

        truncate: {
            true: "truncate",
        },
    },
    defaultVariants: {
        text: "base",
        font: "sans",
        weight: "normal",
        align: "left",
        transform: "normal-case",
        nowrap: false,
        tracking: "normal",
        select: false,
    },
});

