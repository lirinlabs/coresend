import type { ReactNode } from "react";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import {
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from "@/components/ui/tooltip";

interface ActionIconProps {
    icon: ReactNode;
    tooltip: string;
    title: string;
    description: string;
    actionText: string;
    onAction: () => void;
    actionVariant?: "default" | "destructive";
    iconClassName?: string;
}

export const ActionIcon = ({
    icon,
    tooltip,
    title,
    description,
    actionText,
    onAction,
    actionVariant = "default",
}: ActionIconProps) => {
    return (
        <AlertDialog>
            <Tooltip>
                <TooltipTrigger asChild>
                    <button type="button">
                        <AlertDialogTrigger asChild>
                            <div>{icon}</div>
                        </AlertDialogTrigger>
                    </button>
                </TooltipTrigger>
                <TooltipContent>{tooltip}</TooltipContent>
            </Tooltip>
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>{title}</AlertDialogTitle>
                    <AlertDialogDescription>
                        {description}
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                        onClick={onAction}
                        className={
                            actionVariant === "destructive"
                                ? "bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                : ""
                        }
                    >
                        {actionText}
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    );
};
