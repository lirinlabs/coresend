import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import Typography from "../Typography/typography"
import { cn } from "@/lib/utils"

interface AccountIndicatorProps {
  index: number
  address: string
  isSelected: boolean
  onClick: () => void
}

export const AccountIndicator = ({
  index,
  address,
  isSelected,
  onClick,
}: AccountIndicatorProps) => {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          type="button"
          onClick={onClick}
          className={cn(
            "w-8 h-8 border rounded flex items-center justify-center cursor-pointer transition-colors",
            isSelected
              ? "border-primary bg-primary/10 text-primary"
              : "border-border text-muted-foreground hover:bg-secondary"
          )}
        >
          <Typography as="span" text="xs" font="mono">
            {index + 1}
          </Typography>
        </button>
      </TooltipTrigger>
      <TooltipContent side="right">{address}</TooltipContent>
    </Tooltip>
  )
}
