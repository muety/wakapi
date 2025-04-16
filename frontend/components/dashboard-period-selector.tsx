"use client";

import { format, subDays } from "date-fns";
import { startCase } from "lodash";
import { Calendar, CalendarArrowDownIcon } from "lucide-react";
import Link from "next/link";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  buildQueryForRangeQuery,
  cn,
  DashboardRangeQuery,
  getSelectedPeriodLabel,
} from "@/lib/utils";

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
  const linkToToday = `${baseUrl}/day/${format(new Date(), "yyyy-MM-dd")}`;
  const linkToYesterday = `${baseUrl}/day/${format(subDays(new Date(), 1), "yyyy-MM-dd")}`;
  return (
    <div className="flex items-center gap-3">
      <Link
        href={linkToToday}
        className={cn(
          "inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
          "text-white hover:bg-gray-800 h-9 px-3",
          "border border-gray-700 hover:border-gray-600"
        )}
      >
        <Calendar className="w-4 h-4 mr-2" />
        Today
      </Link>
      <Link
        href={linkToYesterday}
        className={cn(
          "inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
          "text-gray-400 hover:bg-gray-800 h-9 px-3",
          "border border-gray-700 hover:border-gray-600"
        )}
      >
        <CalendarArrowDownIcon className="w-4 h-4 mr-2" />
        Yesterday
      </Link>
      <div className="flex items-center bg-gray-800 rounded-md border border-gray-700 cursor-pointer">
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
          <SelectContent className="bg-gray-800 text-white border-gray-700 cursor-pointer">
            {Object.values(DashboardRangeQuery).map((query, index) => (
              <SelectItem
                key={index}
                value={baseUrl + buildQueryForRangeQuery(query)}
                className="cursor-pointer"
              >
                {startCase(query)}
              </SelectItem>
            ))}
            <SelectItem
              className="cursor-pointer"
              key="today"
              value={linkToToday}
            >
              Today
            </SelectItem>
            <SelectItem
              className="cursor-pointer"
              key="yesterday"
              value={linkToYesterday}
            >
              Yesterday
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  );
}
