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
import { makeCategorySummaryData } from "@/lib/utils";

import { ActivityCategoriesSummaryChart } from "./ActivityCategoriesSummaryChart";
import { EmptyChartWrapper } from "./EmptyChartWrapper";
import { StackedTooltipContentForCategories } from "./StackedTooltipContent";

export interface iProps {
  data: SummariesResponse[];
}

export function ActivityCategoriesChartComponent({
  data: rawSummaries,
}: iProps) {
  const [groupedSummaryData, categoryData] =
    makeCategorySummaryData(rawSummaries);
  return (
    <>
      <ActivityCategoriesSummaryChart
        data={categoryData}
        activityColors={COLORS.categories}
      />
      <ResponsiveContainer width="100%" height={195}>
        <BarChart data={groupedSummaryData}>
          <XAxis
            dataKey="name"
            stroke="#8a8e91"
            tickLine={true}
            axisLine={true}
            className="chart-x-axis"
          />
          <RechartsTooltip
            content={StackedTooltipContentForCategories}
            isAnimationActive={false}
            wrapperStyle={{ zIndex: 300 }}
          />
          {Object.keys(categoryData).map((category: string, index: number) => (
            <Bar
              key={index + category}
              dataKey={category}
              fill={COLORS.categories[category]}
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

export function ActivityCategoriesChart({ data }: iProps) {
  console.log("data", data);
  const totalSeconds = data
    .map((d) => d.grand_total.total_seconds || 0)
    .reduce((a, b) => a + b, 0);
  return (
    <EmptyChartWrapper hasData={totalSeconds > 0}>
      <ActivityCategoriesChartComponent data={data} />
    </EmptyChartWrapper>
  );
}
