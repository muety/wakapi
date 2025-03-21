"use client";

import {
  Bar,
  BarChart,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
} from "recharts";

import { COLORS } from "@/lib/constants";
import { SummariesResponse } from "@/lib/types";
import { makeCategorySummaryDataForWeekdays } from "@/lib/utils";

import { DurationTooltip } from "./DurationTooltip";
import { EmptyChartWrapper } from "./EmptyChartWrapper";
import { StackedTooltipContent } from "./StackedTooltipContent";

export interface iProps {
  data: SummariesResponse[];
  durationSubtitle?: string;
}

export function WeekdaysBarChartComponent({ data, durationSubtitle }: iProps) {
  const [groupedSummaryData, categoryData] =
    makeCategorySummaryDataForWeekdays(data);
  return (
    <>
      <DurationTooltip title="WEEKDAYS" subtitle={durationSubtitle} />
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

export function WeekdaysBarChart({ data }: iProps) {
  const totalSeconds = data
    .map((d) => d.grand_total.total_seconds || 0)
    .reduce((a, b) => a + b, 0);
  return (
    <EmptyChartWrapper hasData={totalSeconds > 0}>
      <WeekdaysBarChartComponent data={data} />
    </EmptyChartWrapper>
  );
}
