"use client";

import {
  Bar,
  BarChart,
  ResponsiveContainer,
  XAxis,
  Tooltip as RechartsTooltip,
} from "recharts";
import { makeCategorySummaryDataForWeekdays } from "@/lib/utils";
import { StackedTooltipContent } from "./StackedTooltipContent";
import { SummariesResponse } from "@/lib/types";
import { COLORS } from "@/lib/constants";
import { DurationTooltip } from "./DurationTooltip";

export interface iProps {
  data: SummariesResponse[];
  durationSubtitle?: string;
}

export function WeekdaysBarChart({ data, durationSubtitle }: iProps) {
  const [groupedSummaryData, categoryData] =
    makeCategorySummaryDataForWeekdays(data);
  return (
    <>
      <DurationTooltip title="Weekdays" subtitle={durationSubtitle} />
      <ResponsiveContainer width="100%" height={190}>
        <BarChart data={groupedSummaryData}>
          <XAxis
            dataKey="name"
            stroke="#888888"
            fontSize={12}
            tickLine={false}
            axisLine={false}
          />
          <RechartsTooltip
            content={StackedTooltipContent as any}
            cursor={false}
          />
          {Object.keys(categoryData).map((category, index) => (
            <Bar
              key={index + category}
              dataKey={category}
              fill={COLORS.categories[category]}
              stackId="weekdays_summary"
              radius={
                index === Object.keys(categoryData).length - 1
                  ? [4, 4, 0, 0]
                  : [0, 0, 0, 0]
              }
              width={56}
            />
          ))}
        </BarChart>
      </ResponsiveContainer>
    </>
  );
}
