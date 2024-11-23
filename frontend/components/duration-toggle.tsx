"use client";

import { useState } from "react";
import { CalendarIcon, ChevronDownIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

type DurationOption = {
  label: string;
  value: string;
  tooltip: string;
};

const mainOptions: DurationOption[] = [
  { label: "7D", value: "7D", tooltip: "Last 7 Days" },
  { label: "7DY", value: "7DY", tooltip: "Last 7 Days from Yesterday" },
  { label: "14D", value: "14D", tooltip: "Last 14 Days" },
  { label: "30D", value: "30D", tooltip: "Last 30 Days" },
];

const moreOptions: DurationOption[] = [
  { label: "TM", value: "TM", tooltip: "This Month" },
  { label: "TW", value: "TW", tooltip: "This Week" },
  { label: "LW", value: "LW", tooltip: "Last Week" },
  { label: "LM", value: "LM", tooltip: "Last Month" },
  { label: "Custom", value: "CUSTOM", tooltip: "Select Custom Date Range" },
];

export function DurationToggle({
  onChange,
}: {
  onChange: (value: string) => void;
}) {
  const [activeDuration, setActiveDuration] = useState<string>("7D");

  const handleDurationChange = (value: string) => {
    setActiveDuration(value);
    onChange(value);
  };

  return (
    <TooltipProvider>
      <div className="flex items-center space-x-2 bg-muted p-1 rounded-lg">
        {mainOptions.map((option) => (
          <Tooltip key={option.value}>
            <TooltipTrigger asChild>
              <Button
                variant={activeDuration === option.value ? "default" : "ghost"}
                size="sm"
                onClick={() => handleDurationChange(option.value)}
                className={`${
                  activeDuration === option.value
                    ? "bg-background text-foreground shadow-sm"
                    : ""
                } transition-all duration-200 ease-in-out`}
              >
                {option.label}
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>{option.tooltip}</p>
            </TooltipContent>
          </Tooltip>
        ))}
        <DropdownMenu>
          <Tooltip>
            <TooltipTrigger asChild>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className={`${
                    moreOptions.some((opt) => opt.value === activeDuration)
                      ? "bg-background text-foreground shadow-sm"
                      : ""
                  } transition-all duration-200 ease-in-out`}
                >
                  More <ChevronDownIcon className="ml-1 h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
            </TooltipTrigger>
            <TooltipContent>
              <p>Additional duration options</p>
            </TooltipContent>
          </Tooltip>
          <DropdownMenuContent>
            {moreOptions.map((option) => (
              <DropdownMenuItem
                key={option.value}
                onClick={() => handleDurationChange(option.value)}
              >
                <span>{option.label}</span>
                <span className="ml-auto text-muted-foreground text-xs">
                  {option.tooltip}
                </span>
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </TooltipProvider>
  );
}
