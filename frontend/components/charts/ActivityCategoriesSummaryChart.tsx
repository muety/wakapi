"use client";
import { ResponsiveContainer } from "recharts";

import { BarChart, Bar, XAxis, YAxis, Tooltip } from "recharts";
import { StackedTooltipContentForCategories } from "./StackedTooltipContent";

export const getPercent = (value: number, total: number) => {
  const ratio = total > 0 ? value / total : 0;

  return toPercent(ratio, 2);
};

const toPercent = (decimal: number, fixed = 0) => {
  return `${(decimal * 100).toFixed(fixed)}%`;
};

interface ActivityCategoriesSummaryChartProps {
  data: Record<string, any>;
  activityColors: Record<string, string>;
}

export function ActivityCategoriesSummaryChart({
  data,
  activityColors,
}: ActivityCategoriesSummaryChartProps) {
  return (
    <ResponsiveContainer width="100%" height={30}>
      <BarChart
        height={30}
        data={[data]}
        stackOffset="expand"
        margin={{ top: 0, right: 0, left: 0, bottom: 0 }}
        layout="vertical"
      >
        <XAxis type="number" hide />
        <YAxis type="category" hide />
        <Tooltip
          wrapperStyle={{ zIndex: 300 }}
          content={StackedTooltipContentForCategories as any}
        />
        {Object.keys(data).map((key) => (
          <Bar
            key={key}
            dataKey={key}
            stackId="1"
            stroke={activityColors[key]}
            fill={activityColors[key]}
          />
        ))}
      </BarChart>
    </ResponsiveContainer>
  );
}
