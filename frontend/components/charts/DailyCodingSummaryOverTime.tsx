"use client";

import { useTheme } from "next-themes";
import { useState } from "react";
import {
  Bar,
  ComposedChart,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
} from "recharts";

import { SAMPLE_COLORS } from "@/lib/constants";
import { SummariesResponse } from "@/lib/types";
import {
  getUniqueProjects,
  prepareDailyCodingData,
  transparentize,
} from "@/lib/utils";

import { StackedTooltipContent } from "./StackedTooltipContent";

export interface iProps {
  data: SummariesResponse[];
}

export function DailyCodingSummaryOverTime({ data }: iProps) {
  const { theme } = useTheme();
  const lineColor = theme === "dark" ? "#ffffff" : "#000";
  // Resolve whatever fuckery is afoot here.
  // Normalizing seems to bring us closer to an optimum solution but has its own
  // pitfals? It just looks nicer but the labels are fucked
  const chartData = data.map(prepareDailyCodingData);
  const uniqueProjects = getUniqueProjects(chartData);

  const customUniqueProjects: Record<string, string> = uniqueProjects.reduce(
    (prev, cur, index) => ({
      ...prev,
      [cur]: SAMPLE_COLORS[index % SAMPLE_COLORS.length],
    }),
    {}
  );

  const [focusDataIndex, setFocusDataIndex] = useState<number | null>(null);
  return (
    <ResponsiveContainer
      width="100%"
      height={220}
      style={{ background: "transparent" }}
    >
      <ComposedChart
        style={{ background: "transparent" }}
        width={500}
        height={220}
        data={chartData}
        barGap={"100%"}
        barCategoryGap={3}
        margin={{
          top: 20,
          right: 0,
          left: 0,
          bottom: 5,
        }}
      >
        <XAxis
          dataKey="name"
          stroke="#8a8e91"
          tickLine={true}
          axisLine={true}
          className="chart-x-axis"
        />
        <Tooltip
          wrapperStyle={{ zIndex: 500 }}
          content={StackedTooltipContent as any}
          cursor={false}
        />
        {Object.entries(customUniqueProjects).map(([dataKey, color]) => (
          <Bar
            key={dataKey}
            dataKey={dataKey}
            fill={transparentize(color, 0.83)}
            stroke={color}
            stackId={"groot"}
            strokeWidth={1.3}
            onMouseMove={(e: any) => {
              if (e.activeTooltipIndex !== focusDataIndex) {
                setFocusDataIndex(e.activeTooltipIndex);
              }
            }}
            onMouseLeave={() => setFocusDataIndex(null)}
          />
        ))}
        <Line
          type="natural"
          dataKey="total"
          stroke={lineColor}
          dot={false}
          strokeWidth={1.45}
          tooltipType="none"
        />
      </ComposedChart>
    </ResponsiveContainer>
  );
}
