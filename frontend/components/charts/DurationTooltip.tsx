import { QuestionMarkCircledIcon } from "@radix-ui/react-icons";
import { useSearchParams } from "next/navigation";

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { getSelectedPeriodLabel } from "@/lib/utils";

interface iProps {
  title: string;
  subtitle?: string;
}

export function DurationTooltip({ title, subtitle }: iProps) {
  const params = useSearchParams();
  const start = params.get("start");
  const end = params.get("end");
  const durationText = getSelectedPeriodLabel(
    end && start ? { end, start } : {}
  );
  return (
    <div className="chart-box-title">
      {title}
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <QuestionMarkCircledIcon className="cursor-pointer hover:opacity-70" />
          </TooltipTrigger>
          <TooltipContent>
            <p>
              {subtitle} {durationText}
            </p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}
