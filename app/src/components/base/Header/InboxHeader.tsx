import TrashIcon from "@/components/ui/trash-icon";
import { Logo } from "../Logo/Logo";
import { ModeToggle } from "../ModeToggle/ModeToggle";
import CopyIcon from "@/components/ui/copy-icon";
import FlameIcon from "@/components/ui/flame-icon";
import Typography from "../Typography/typography";
import { ActionIcon } from "../ActionIcon";
import { SettingsMenu } from "../SettingsMenu/SettingsMenu";

export const InboxHeader = () => {
    return (
        <header className="w-full p-4 border-b border-border">
            <div className="max-w-7xl mx-auto flex items-center justify-between">
                <Logo navigate />

                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2 bg-secondary p-1.5 -my-2 pr-2 rounded-xs">
                        <Typography
                            color="muted"
                            text="xs"
                            font="mono"
                            weight="semibold"
                        >
                            4df...432C@coresend.io
                        </Typography>
                        <CopyIcon className="text-muted-foreground hover:text-primary transition-colors h-3 w-3" />
                    </div>
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
                        iconClassName="text-muted-foreground hover:text-primary transition-colors h-4 w-4"
                    />

                    <SettingsMenu />

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
