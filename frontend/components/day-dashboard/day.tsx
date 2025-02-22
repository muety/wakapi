"use client";

import { Group } from "@visx/group";
import { scaleLinear, scaleTime } from "@visx/scale";
import { Bar, Line } from "@visx/shape";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type React from "react";
import { useMemo } from "react";

interface RawTimeEntry {
  time: number;
  project: string;
  duration: number;
  color: string | null;
}

interface RawData {
  data: RawTimeEntry[];
  start: string;
  end: string;
  timezone: string;
}

interface ProcessedActivity {
  id: string;
  start: Date;
  end: Date;
  project: string;
  duration: number;
}

interface ProjectGroup {
  name: string;
  activities: ProcessedActivity[];
  totalDuration: number;
}

interface TimeTrackingProps {
  width: number;
  margin?: { top: number; right: number; bottom: number; left: number };
  data: RawData;
}

const defaultMargin = { top: 140, right: 40, bottom: 40, left: 220 };
const ROW_HEIGHT = 50;
const LANE_HEIGHT = 25;

const TimeTrackingVisualization: React.FC<TimeTrackingProps> = ({
  width,
  margin = defaultMargin,
  data: rawData,
}) => {
  const { timeScale, yScale, projects, totalTime, calculatedHeight } =
    useMemo(() => {
      const startDate = new Date(rawData.start);
      const endDate = new Date(rawData.end);

      const activities: ProcessedActivity[] = rawData.data.map((item) => ({
        id: item.time.toString(),
        start: new Date(item.time * 1000),
        end: new Date((item.time + item.duration) * 1000),
        project: item.project,
        duration: item.duration,
      }));

      const projectGroups: { [key: string]: ProcessedActivity[] } = {};
      activities.forEach((activity) => {
        if (!projectGroups[activity.project]) {
          projectGroups[activity.project] = [];
        }
        projectGroups[activity.project].push(activity);
      });

      const projects: ProjectGroup[] = Object.entries(projectGroups)
        .map(([name, activities]) => ({
          name,
          activities,
          totalDuration: activities.reduce((sum, act) => sum + act.duration, 0),
        }))
        .sort((a, b) => b.totalDuration - a.totalDuration);

      const totalTime = projects.reduce(
        (sum, project) => sum + project.totalDuration,
        0
      );
      const calculatedHeight =
        projects.length * ROW_HEIGHT + margin.top + margin.bottom;

      const timeScale = scaleTime({
        domain: [startDate, endDate],
        range: [margin.left, width - margin.right],
      });

      const yScale = scaleLinear({
        domain: [0, projects.length],
        range: [margin.top, calculatedHeight - margin.bottom],
      });

      return { timeScale, yScale, projects, totalTime, calculatedHeight };
    }, [width, margin, rawData]);

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}:${minutes.toString().padStart(2, "0")}`;
  };

  const totalHours = Math.floor(totalTime / 3600);
  const totalMinutes = Math.floor((totalTime % 3600) / 60);

  const currentDate = new Date();
  const formattedDate = currentDate.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  const getHourLabel = (date: Date) => {
    const hour = date.getHours();
    if (hour === 0) return "12a";
    if (hour === 12) return "12p";
    return `${hour % 12}${hour < 12 ? "a" : "p"}`;
  };

  return (
    <div className="relative bg-[rgb(18,18,18)] rounded-xl p-4">
      {/* Header */}
      <div className="absolute top-4 left-0 right-0 flex items-center justify-center gap-3 px-4">
        <ChevronLeft className="w-5 h-5 text-[#3b82f6] cursor-pointer hover:text-white" />
        <div className="flex items-center gap-2">
          <span className="text-xl font-medium text-white">
            {totalHours} hrs {totalMinutes} mins
          </span>
          <span className="text-xl text-[#3b82f6]">on</span>
          <span className="text-xl text-[#3b82f6] border-b border-dashed border-[#3b82f6]">
            {formattedDate}
          </span>
        </div>
        <ChevronRight className="w-5 h-5 text-[#3b82f6] cursor-pointer hover:text-white" />
      </div>

      <svg width={width} height={calculatedHeight}>
        {/* Surrounding box lines */}
        <Line
          from={{ x: margin.left, y: margin.top }}
          to={{ x: margin.left, y: calculatedHeight - margin.bottom }}
          stroke="rgba(255,255,255,0.1)"
          strokeWidth={1}
        />
        <Line
          from={{ x: width - margin.right, y: margin.top }}
          to={{ x: width - margin.right, y: calculatedHeight - margin.bottom }}
          stroke="rgba(255,255,255,0.1)"
          strokeWidth={1}
        />

        {/* Time axis with ticks below labels */}
        <Group top={margin.top - 50}>
          {timeScale.ticks(24).map((date, i) => {
            const x = timeScale(date);
            const label = getHourLabel(date);
            const isEdge = i === 0 || i === 23;

            return (
              <g key={i}>
                <text
                  x={x}
                  y={30} // Position text above the tick
                  textAnchor={isEdge ? (i === 0 ? "start" : "end") : "middle"}
                  fill="#6b7280"
                  fontSize={12}
                  fontFamily="monospace"
                >
                  {label}
                </text>
                <Line
                  from={{ x, y: 40 }} // Start tick below text
                  to={{ x, y: 50 }} // End tick at top of first row
                  stroke="rgba(255,255,255,0.1)"
                  strokeWidth={1}
                />
              </g>
            );
          })}
        </Group>

        {/* Project lanes */}
        {projects.map((project, index) => {
          const yPos = yScale(index);
          return (
            <Group key={project.name} top={yPos}>
              {/* First row top line to align with tick endings */}
              {index === 0 && (
                <Line
                  from={{ x: margin.left, y: 0 }}
                  to={{ x: width - margin.right, y: 0 }}
                  stroke="rgba(255,255,255,0.1)"
                  strokeWidth={1}
                />
              )}

              {/* Row background */}
              <rect
                x={margin.left}
                y={0}
                width={width - margin.left - margin.right}
                height={ROW_HEIGHT}
                fill="rgba(255,255,255,0.03)"
              />

              {/* Project label */}
              <text
                x={margin.left - 20}
                y={ROW_HEIGHT / 2 + 10}
                fill="#fff"
                textAnchor="end"
                fontSize={14}
                fontFamily="monospace"
              >
                {`${project.name} ${formatDuration(project.totalDuration)}`}
              </text>

              {/* Activity bars */}
              {project.activities.map((activity) => (
                <Bar
                  key={activity.id}
                  x={timeScale(activity.start)}
                  y={(ROW_HEIGHT - LANE_HEIGHT) / 2}
                  width={Math.max(
                    2,
                    timeScale(activity.end) - timeScale(activity.start)
                  )}
                  height={LANE_HEIGHT}
                  fill="#3b82f6"
                  rx={2}
                  style={{ cursor: "pointer" }}
                />
              ))}

              {/* Bottom row line */}
              <Line
                from={{ x: margin.left, y: ROW_HEIGHT }}
                to={{ x: width - margin.right, y: ROW_HEIGHT }}
                stroke="rgba(255,255,255,0.1)"
                strokeWidth={1}
              />
            </Group>
          );
        })}
      </svg>
    </div>
  );
};

export default TimeTrackingVisualization;
