import { InboxHeader } from "@/components/base/Header/InboxHeader";

const Inbox = () => {
    return (
        <div className="w-full h-dvh flex flex-col">
            {/* Invisible, SEO purpose only */}
            <h1 className="sr-only">Stateless temporary email.</h1>

            <InboxHeader />
            <div className="max-w-7xl mx-auto w-full flex-1 flex flex-col justify-center items-center">
                <div className="flex flex-col items-center gap-4">test</div>
            </div>
        </div>
    );
};

export default Inbox;
