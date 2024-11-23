"use client";

import {
  Bar,
  BarChart,
  ResponsiveContainer,
  XAxis,
  Tooltip as RechartsTooltip,
} from "recharts";
import { COLORS } from "@/lib/constants";
import { SummariesResponse } from "@/lib/types";
import { StackedTooltipContentForCategories } from "./StackedTooltipContent";
import { ActivityCategoriesSummaryChart } from "./ActivityCategoriesSummaryChart";
import { makeCategorySummaryData } from "@/lib/utils";

export interface iProps {
  data: SummariesResponse[];
}

export function ActivityCategoriesChart({ data: rawSummaries }: iProps) {
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
            content={StackedTooltipContentForCategories as any}
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
