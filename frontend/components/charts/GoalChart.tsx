"use client";

import {
  Bar,
  Cell,
  ComposedChart,
  Label,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { GoalChartData } from "@/lib/types";
import { convertSecondsToHours } from "@/lib/utils";

import { StackedTooltipContent } from "./StackedTooltipContent";

interface iProps {
  data: GoalChartData[];
  direction: "more" | "less";
}

export function GoalChart({ data, direction }: iProps) {
  const targetMet = (entry: GoalChartData) => {
    if (direction === "more") {
      return entry.actual_seconds > entry.goal_seconds;
    }
    return entry.actual_seconds < entry.goal_seconds;
  };

  return (
    <ResponsiveContainer width="100%" height={230}>
      <ComposedChart
        style={{ background: "transparent" }}
        height={230}
        data={data}
      >
        <XAxis
          dataKey="range.text"
          stroke="#888888"
          fontSize={12}
          tickLine={true}
          axisLine={true}
        />
        <YAxis
          tickFormatter={convertSecondsToHours}
          className="chart-x-axis"
          tickLine={true}
          axisLine={true}
        >
          <Label angle={-90} />
        </YAxis>
        <Tooltip content={StackedTooltipContent as any} cursor={false} />

        <Bar dataKey="actual_seconds" fill="#48d875" radius={[5, 5, 0, 0]}>
          {data.map((entry, index) => {
            const color = targetMet(entry) ? "#48d875" : "#ff5252";
            return <Cell fill={color} key={index} />;
          })}
        </Bar>
        <Line
          type="natural"
          dataKey="goal_seconds"
          stroke="red"
          dot={false}
          strokeWidth={1.1}
          tooltipType="none"
          fill={"red"}
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}
