"use client";

import { ChevronDownIcon } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

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
      <div className="flex items-center space-x-2 rounded-lg bg-muted p-1">
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
                  More <ChevronDownIcon className="ml-1 size-4" />
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
                <span className="ml-auto text-xs text-muted-foreground">
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
