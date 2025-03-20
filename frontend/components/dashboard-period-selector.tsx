"use client";

import { format, subDays } from "date-fns";
import { startCase } from "lodash";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  buildQueryForRangeQuery,
  DashboardRangeQuery,
  getSelectedPeriodLabel,
} from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export function DashboardPeriodSelector({
  searchParams,
  baseUrl = "/dashboard",
}: {
  searchParams: Record<string, any>;
  baseUrl?: string;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger>
        <span
          className="cursor-pointer"
          style={{ borderBottom: "1px solid black", color: "#36558b" }}
        >
          {getSelectedPeriodLabel(searchParams || {})}
        </span>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="period-date-picker w-56">
        <DropdownMenuItem
          className="cursor-pointer hover:bg-primary hover:text-white"
          asChild
        >
          <a href={`${baseUrl}/day?date=${format(new Date(), "yyyy-MM-dd")}`}>
            Today
          </a>
        </DropdownMenuItem>
        <DropdownMenuItem
          className="cursor-pointer hover:bg-primary hover:text-white"
          asChild
        >
          <a
            href={`${baseUrl}/day?date=${format(
              subDays(new Date(), 1),
              "yyyy-MM-dd"
            )}`}
          >
            Yesterday
          </a>
        </DropdownMenuItem>
        {Object.values(DashboardRangeQuery).map((query, index) => (
          <DropdownMenuItem
            className="period-date-picker cursor-pointer hover:bg-primary hover:text-white"
            key={index}
          >
            <a href={baseUrl + buildQueryForRangeQuery(query)}>
              {startCase(query)}
            </a>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function DashboardPeriodSelectorV2({
  searchParams,
  baseUrl = "/dashboard",
}: {
  searchParams: Record<string, any>;
  baseUrl?: string;
}) {
  return (
    <div className="flex items-center gap-3">
      <div className="flex items-center bg-gray-800 rounded-md border border-gray-700">
        <Select
          onValueChange={(value) => {
            window.location.href = value;
          }}
        >
          <SelectTrigger className="w-[180px] bg-transparent border-0 text-white focus:ring-0 focus:ring-offset-0">
            <SelectValue
              placeholder={getSelectedPeriodLabel(searchParams || {})}
            />
          </SelectTrigger>
          <SelectContent className="bg-gray-800 text-white border-gray-700">
            {Object.values(DashboardRangeQuery).map((query, index) => (
              <SelectItem
                key={index}
                value={baseUrl + buildQueryForRangeQuery(query)}
              >
                {startCase(query)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </div>
  );
}
