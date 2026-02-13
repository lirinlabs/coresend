import { Plus } from "@phosphor-icons/react"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"

interface AddAccountButtonProps {
  onClick: () => void
}

export const AddAccountButton = ({ onClick }: AddAccountButtonProps) => {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          type="button"
          onClick={onClick}
          className="w-8 h-8 border border-dashed border-border rounded flex items-center justify-center hover:bg-secondary hover:border-solid transition-colors"
        >
          <Plus weight="bold" className="w-4 h-4 text-muted-foreground" />
        </button>
      </TooltipTrigger>
      <TooltipContent side="right">Derive new address</TooltipContent>
    </Tooltip>
  )
}
