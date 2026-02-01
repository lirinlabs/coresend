import TrashIcon from "@/components/ui/trash-icon";
import { Logo } from "../Logo/Logo";
import { ModeToggle } from "../ModeToggle/ModeToggle";
import GearIcon from "@/components/ui/gear-icon";
import CopyIcon from "@/components/ui/copy-icon";
import FlameIcon from "@/components/ui/flame-icon";
import Typography from "../Typography/typography";
import { ActionIcon } from "../ActionIcon";

export const InboxHeader = () => {
    return (
        <header className="w-full py-4 border-b border-border">
            <div className="max-w-7xl mx-auto flex items-center justify-between">
                <Logo />
                <div className="flex items-center gap-2 bg-secondary p-1.5 -my-2 pr-2 rounded-xs">
                    <Typography
                        color="secondary-foreground"
                        text="xs"
                        font="mono"
                        weight="semibold"
                    >
                        4df...432C@coresend.io
                    </Typography>
                    <CopyIcon className="text-muted-foreground hover:text-primary transition-colors h-3 w-3" />
                </div>
                <div className="flex items-center gap-4">
                    <ModeToggle />

                    <ActionIcon
                        icon={
                            <TrashIcon className="text-muted-foreground hover:text-primary transition-colors h-4 w-4 " />
                        }
                        tooltip="Wipe inbox"
                        title="Wipe Inbox"
                        description="This will permanently delete all emails in the current inbox."
                        actionText="Wipe All"
                        onAction={() => console.log("Wipe inbox")}
                        actionVariant="destructive"
                    />

                    <ActionIcon
                        icon={
                            <FlameIcon className="text-muted-foreground hover:text-primary transition-colors h-4 w-4" />
                        }
                        tooltip="Burn inbox"
                        title="Burn Inbox"
                        description="This will permanently delete the entire inbox including the address."
                        actionText="Burn Now"
                        onAction={() => console.log("Burn inbox")}
                        actionVariant="destructive"
                        iconClassName="text-muted-foreground hover:text-primary transition-colors h-4 w-4"
                    />

                    <GearIcon className="text-muted-foreground hover:text-primary transition-colors h-4 w-4" />

                    <Typography
                        text="xs"
                        font="mono"
                        tracking="tight"
                        transform="uppercase"
                        color="muted"
                    >
                        LOGOUT
                    </Typography>
                </div>
            </div>
        </header>
    );
};
