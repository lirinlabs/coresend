import { typography } from './config';
import type { VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';
import React from 'react';

type ElementType =
    | 'p'
    | 'span'
    | 'div'
    | 'h1'
    | 'h2'
    | 'h3'
    | 'h4'
    | 'h5'
    | 'h6'
    | 'label';

export interface TypographyProps
    extends
        Omit<React.HTMLAttributes<HTMLElement>, 'color'>,
        VariantProps<typeof typography> {
    as?: ElementType;
}

const Typography = React.forwardRef<HTMLElement, TypographyProps>(
    (
        {
            className,
            as: Component = 'p',
            text,
            textLg,
            color,
            font,
            weight,
            align,
            transform,
            tracking,
            select,
            nowrap,
            ellipsis,
            truncate,
            ...props
        },
        ref,
    ) => {
        return React.createElement(
            Component,
            {
                ref,
                className: cn(
                    typography({
                        text,
                        textLg,
                        color,
                        font,
                        weight,
                        align,
                        transform,
                        tracking,
                        select,
                        nowrap,
                        ellipsis,
                        truncate,
                    }),
                    className,
                ),
                ...props,
            },
            props.children,
        );
    },
);

Typography.displayName = 'Typography';

export default Typography;
