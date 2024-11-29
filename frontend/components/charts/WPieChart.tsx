"use client";

import React, { useCallback, useState } from "react";
import {
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Sector,
  Tooltip as RechartsTooltip,
} from "recharts";

import { useMediaQuery } from "@/hooks/use-media-query";
import {
  convertSecondsToHoursAndMinutes,
  getEntityColor,
  getMachineColor,
} from "@/lib/utils";

import { DurationTooltip } from "./DurationTooltip";

export interface WPieChartDataItem {
  key: string;
  total: number;
  // color: string;
}

export interface WPieChartProps {
  data: WPieChartDataItem[];
  title: string;
  innerRadius?: number;
  colorNamespace: string;
  durationSubtitle?: string;
}

function legendFormatter(
  _: string,
  entry: { payload: { total: number; percent: number; key: string } },
  index: number
) {
  return (
    <text
      fill="#8a8e91"
      stroke="#8a8e91"
      style={{
        fontSize: "9px",
        color: "#8a8e91",
        marginRight: "58px",
      }}
      key={index}
    >
      {entry.payload.key} -{" "}
      {convertSecondsToHoursAndMinutes(entry.payload.total)} (
      {(entry.payload.percent * 100).toFixed(0)}%)
    </text>
  );
}

export function WPieChart({
  data,
  innerRadius = 0,
  title,
  durationSubtitle,
  colorNamespace,
}: WPieChartProps) {
  const hideLegend = useMediaQuery("only screen and (max-width : 576px)");

  const [, setActiveIndex] = useState(null);
  const onMouseOver = useCallback(
    (_data: any, index: React.SetStateAction<null>) => {
      setActiveIndex(index);
    },
    []
  );
  const onMouseLeave = useCallback(() => {
    setActiveIndex(null);
  }, []);

  const renderActiveShape = (props: any) => {
    const RADIAN = Math.PI / 180;
    const {
      cx,
      cy,
      innerRadius,
      outerRadius,
      startAngle,
      endAngle,
      midAngle,
      fill,
    } = props;
    const sin = Math.sin(-RADIAN * midAngle);
    const cos = Math.cos(-RADIAN * midAngle);
    const sx = cx + (outerRadius - 77) * cos;
    const sy = cy + (outerRadius - 77) * sin;
    return (
      <Sector
        cx={sx}
        cy={sy}
        innerRadius={innerRadius}
        outerRadius={outerRadius}
        startAngle={startAngle}
        endAngle={endAngle}
        fill={fill}
      />
    );
  };

  function renderCustomizedLabel(rawData: any) {
    const RADIAN = Math.PI / 180;
    const { cx, cy, midAngle, innerRadius, outerRadius, index, percent } =
      rawData;
    const radius = innerRadius + (outerRadius - innerRadius) * 0.5;
    const x = cx + radius * Math.cos(-midAngle * RADIAN);
    const y = cy + radius * Math.sin(-midAngle * RADIAN);

    return (
      <text
        x={x}
        y={y}
        fill="white"
        textAnchor={x > cx ? "start" : "end"}
        dominantBaseline="central"
        style={{ fontSize: "13px" }}
        key={index}
        alignmentBaseline="central"
      >
        {percent >= 0.15 ? data[index].key : ""}
      </text>
    );
  }

  const CustomizedTooltip = (props: any) => {
    if (props.payload.length > 0) {
      const payload: any = props.payload[0];
      const percent =
        payload.value / data.reduce((pre, cur) => pre + cur.total, 0);
      return (
        <span
          style={{
            fontSize: "11px",
            color: "black",
            backgroundColor: "white",
            padding: 5,
            border: "1px solid black",
          }}
        >
          {payload.payload.key} -{" "}
          {convertSecondsToHoursAndMinutes(payload.value)} (
          {(percent * 100).toFixed(0)}%)
        </span>
      );
    }
    return null;
  };
  return (
    <>
      <DurationTooltip title={title} subtitle={durationSubtitle} />
      <ResponsiveContainer
        width="100%"
        height={200}
        className="overflow-y-scroll"
      >
        <PieChart>
          {!hideLegend && (
            <Legend
              align="right"
              verticalAlign="top"
              className="hidden md:block"
              layout="vertical"
              style={{ marginRight: "100px" }}
              iconType="circle"
              wrapperStyle={{
                height: "80%",
                overflow: "auto",
              }}
              max={100}
              cy="20%"
              cx="10%"
              iconSize={10}
              formatter={legendFormatter as any}
            ></Legend>
          )}
          <Pie
            dataKey="total"
            isAnimationActive={true}
            data={data}
            cx="50%"
            cy="42%"
            width="100%"
            alignmentBaseline="hanging"
            outerRadius={80}
            innerRadius={innerRadius}
            fill="#8884d8"
            label={renderCustomizedLabel}
            labelLine={false}
            activeShape={renderActiveShape as any}
            onMouseOver={onMouseOver as any}
            onMouseLeave={onMouseLeave}
          >
            {(data || []).map((entry, index) => (
              <Cell
                key={`cell-${index}`}
                fill={
                  colorNamespace != "machines"
                    ? getEntityColor(colorNamespace, entry.key)
                    : getMachineColor(entry.key)
                }
                stroke="black"
                height={"100%"}
                width={"100%"}
              />
            ))}
          </Pie>
          <RechartsTooltip content={<CustomizedTooltip />} />
        </PieChart>
      </ResponsiveContainer>
    </>
  );
}
