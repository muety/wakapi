"use client";

import React from "react";
import {
  Cell,
  Label,
  Pie,
  PieChart,
  ResponsiveContainer,
  Sector,
  Tooltip as RechartsTooltip,
} from "recharts";

import { DailyAverage, SummariesResponse } from "@/lib/types";
import { convertSecondsToHoursAndMinutes } from "@/lib/utils";

import { EmptyChartWrapper } from "./EmptyChartWrapper";

export interface WPieChartDataItem {
  name: string;
  value: number;
  color: string;
}

export interface WPieChartProps {
  data: WPieChartDataItem[];
  title: string;
  innerRadius?: number;
}

export interface WGaugeChartProps {
  data: SummariesResponse[];
  dailyAverage: DailyAverage;
}

export function WGaugeChart({ data, dailyAverage }: WGaugeChartProps) {
  const innerRadius = 45;

  const grandTotal = data
    .map((d) => d.grand_total.total_seconds)
    .reduce((a, b) => a + b, 0);

  const todaysTotal = data[data.length - 1].grand_total.total_seconds;

  const percentageOfDailyAverage = (todaysTotal / dailyAverage.seconds) * 100;
  const change = percentageOfDailyAverage - 100;

  const gaugeData = [
    { value: todaysTotal, color: "hsl(53.25deg 84.36% 64.9%)" },
    { value: dailyAverage.seconds, color: "#e0e0e0" },
  ];

  const hasData = React.useMemo(() => {
    return grandTotal > 0;
  }, [todaysTotal, dailyAverage]);

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
    const { cx, cy, midAngle, innerRadius, outerRadius, index } = rawData;
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
      >
        {/* // TODO: Fix Me please */}
        {/* {percent >= 0.2 ? gaugeData[index].value : ""} */}
      </text>
    );
  }

  const CustomizedTooltip = (props: any) => {
    if (props.payload.length > 0) {
      const payload: any = props.payload[0];
      const percent =
        payload.value / gaugeData.reduce((pre, cur) => pre + cur.value, 0);
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
          {convertSecondsToHoursAndMinutes(payload.value)} (
          {(change < 0 ? change : percent * 100).toFixed(0)}%)
        </span>
      );
    }
    return null;
  };

  function getFillColor(change: number) {
    if (change > 0) {
      return "hsl(120deg 56.1% 40.2%)";
    }

    if (change > -50) {
      return "hsl(53.25deg 84.36% 64.9%)";
    }
    return "hsl(359.66deg 68.5% 49.8%)";
  }

  return (
    <>
      <div className="d-flex">
        <div className="chart-box-title">
          {convertSecondsToHoursAndMinutes(todaysTotal)}
          <span className="chart-box-sub-title">Today</span>
        </div>
      </div>
      <EmptyChartWrapper hasData={hasData} className="h-[150px]">
        <ResponsiveContainer
          width="100%"
          height={160}
          className="flex justify-around align-middle"
        >
          <PieChart>
            <Pie
              dataKey="value"
              isAnimationActive={true}
              data={change > 0 ? [gaugeData[0]] : gaugeData}
              cx="50%"
              cy="75%"
              width="100%"
              alignmentBaseline="hanging"
              outerRadius={85}
              innerRadius={innerRadius}
              fill="#8884d8"
              label={renderCustomizedLabel}
              labelLine={false}
              activeShape={renderActiveShape}
              startAngle={180}
              endAngle={0}
            >
              {gaugeData.map((row, index) => (
                <Cell
                  key={index}
                  fill={index === 0 ? getFillColor(change) : row.color}
                  stroke="black"
                />
              ))}
              <Label
                value={change.toFixed(0) + "%"}
                position="centerBottom"
                offset={-20}
                className="gauge-label"
                fontSize="15px"
                fontWeight="bold"
              />
            </Pie>
            <RechartsTooltip content={<CustomizedTooltip />} />
          </PieChart>
        </ResponsiveContainer>
      </EmptyChartWrapper>
      <div className="d-flex " style={{ marginTop: "-22px" }}>
        <div className="chart-box-title flex flex-col" style={{ gap: 0 }}>
          <span className="text">
            {dailyAverage?.text}
            <span className="chart-box-sub-title">Daily Average</span>
          </span>
          {/* TODO: Support Most Active Day */}
          {/* <span>
            {" "}
            {format(new Date(), "EEE LLL do")}
            <span className="chart-box-sub-title">Most Active Day</span>
          </span> */}
        </div>
      </div>
    </>
  );
}
