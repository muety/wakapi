"use client";

import { Group } from "@visx/group";
import { scaleLinear, scaleTime } from "@visx/scale";
import { Bar, Line } from "@visx/shape";
import { addDays, format, isToday, subDays } from "date-fns";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useRouter } from "next/navigation";
import type React from "react";
import { useEffect, useMemo, useRef, useState } from "react";

import { convertSecondsToHoursAndMinutes } from "@/lib/utils";

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
  data: RawData;
  margin?: { top: number; right: number; bottom: number; left: number };
}

const defaultMargin = { top: 80, right: 40, bottom: 40, left: 220 };
const ROW_HEIGHT = 45;
const LANE_HEIGHT = 33;

interface DateNavigationProps {
  data: RawData;
  totalTime: number;
}

export function DayHeader({ data, totalTime }: DateNavigationProps) {
  const router = useRouter();

  const totalHours = Math.floor(totalTime / 3600);
  const totalMinutes = Math.floor((totalTime % 3600) / 60);

  const currentDate = new Date(data.start);
  const formattedDate = currentDate.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  const DATE_FORMAT = "yyyy-MM-dd";

  const gotoPreviousDay = () => {
    router.push(
      `/dashboard/day/${format(subDays(currentDate, 1), DATE_FORMAT)}`
    );
  };

  const onCurrentDay = isToday(currentDate);

  const gotoNextDay = () => {
    if (onCurrentDay) return;

    const nextDate = addDays(currentDate, 1);
    router.push(`/dashboard/day/${format(nextDate, DATE_FORMAT)}`);
  };

  return (
    <div className="flex items-center justify-center p-4 w-full">
      <button
        onClick={gotoPreviousDay}
        className="flex items-center justify-start size-8 rounded-full border border-blue-500/50 hover:bg-blue-500/10 transition-all"
        aria-label="Previous day"
      >
        <ChevronLeft className="size-8 text-blue-500 hover:text-blue-400 transition-colors" />
      </button>

      <div className="flex items-center mx-5">
        <span className="text-4xl font-bold text-white mr-3">
          {totalHours} hrs {totalMinutes} mins
        </span>
        <span className="text-2xl text-gray-400 mr-3">on</span>
        <span className="text-3xl text-blue-500">{formattedDate}</span>
      </div>

      <button
        onClick={gotoNextDay}
        className="flex items-center justify-end size-8 rounded-full border border-blue-500/50 hover:bg-blue-500/10 transition-all"
        aria-label="Next day"
        disabled={onCurrentDay}
      >
        <ChevronRight className="size-8 text-blue-500 hover:text-blue-400 transition-colors" />
      </button>
    </div>
  );
}

const TimeTrackingVisualization: React.FC<TimeTrackingProps> = ({
  data: rawData,
  margin = defaultMargin,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 });
  const [tooltip, setTooltip] = useState<{
    x: number;
    y: number;
    activity: ProcessedActivity;
  } | null>(null);

  // Effect to measure and update container dimensions
  useEffect(() => {
    if (!containerRef.current) return;

    const updateDimensions = () => {
      if (containerRef.current) {
        const { width } = containerRef.current.getBoundingClientRect();
        setDimensions((prev) => ({ ...prev, width }));
      }
    };

    // Initial measurement
    updateDimensions();

    // Copy the ref value to a local variable - helps with cleanup
    const currentContainer = containerRef.current;

    // Setup resize observer for responsive updates
    const resizeObserver = new ResizeObserver(updateDimensions);
    resizeObserver.observe(currentContainer);

    // Cleanup
    return () => {
      if (currentContainer) {
        resizeObserver.unobserve(currentContainer);
      }
      resizeObserver.disconnect();
    };
  }, []);

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

      // Calculate height based on projects and adjust to container if needed
      const calculatedHeight =
        projects.length * ROW_HEIGHT + margin.top + margin.bottom;

      const timeScale = scaleTime({
        domain: [startDate, endDate],
        range: [margin.left, dimensions.width - margin.right],
      });

      const yScale = scaleLinear({
        domain: [0, projects.length],
        range: [margin.top, calculatedHeight - margin.bottom],
      });

      return { timeScale, yScale, projects, totalTime, calculatedHeight };
    }, [dimensions.width, margin, rawData]);

  // Update height when calculated height changes
  useEffect(() => {
    setDimensions((prev) => ({ ...prev, height: calculatedHeight }));
  }, [calculatedHeight]);

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}:${minutes.toString().padStart(2, "0")}`;
  };

  const getHourLabel = (date: Date) => {
    const hour = date.getHours();
    if (hour === 0) return "12a";
    if (hour === 12) return "12p";
    return `${hour % 12}${hour < 12 ? "a" : "p"}`;
  };

  // Responsive tick generation based on available width
  const tickCount = useMemo(() => {
    // Use fewer ticks on smaller screens
    const width = dimensions.width;
    if (width < 600) return 6; // Every 4 hours
    if (width < 960) return 12; // Every 2 hours
    return 24; // Every hour
  }, [dimensions.width]);

  return (
    <div
      ref={containerRef}
      className="relative bg-[rgb(18,18,18)] rounded-xl p-4 w-full h-full"
    >
      <DayHeader data={rawData} totalTime={totalTime} />

      <svg width="100%" height={dimensions.height}>
        {/* Surrounding box lines */}
        <Line
          from={{ x: margin.left, y: margin.top }}
          to={{ x: margin.left, y: calculatedHeight - margin.bottom }}
          stroke="rgba(255,255,255,0.1)"
          strokeWidth={1}
        />
        <Line
          from={{ x: dimensions.width - margin.right, y: margin.top }}
          to={{
            x: dimensions.width - margin.right,
            y: calculatedHeight - margin.bottom,
          }}
          stroke="rgba(255,255,255,0.1)"
          strokeWidth={1}
        />

        {/* Time axis with ticks below labels - reduced the positioning from margin.top - 50 to margin.top - 25 */}
        <Group top={margin.top - 25}>
          {timeScale.ticks(tickCount).map((date, i) => {
            const x = timeScale(date);
            const label = getHourLabel(date);
            const isEdge = i === 0 || i === tickCount - 1;

            return (
              <g key={i}>
                <text
                  x={x}
                  y={15} // Reduced from 30 to 15
                  textAnchor={isEdge ? (i === 0 ? "start" : "end") : "middle"}
                  fill="#6b7280"
                  fontSize={12}
                  fontFamily="monospace"
                >
                  {label}
                </text>
                <Line
                  from={{ x, y: 20 }} // Reduced from 40 to 20
                  to={{ x, y: 25 }} // Reduced from 50 to 25
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
                  to={{ x: dimensions.width - margin.right, y: 0 }}
                  stroke="rgba(255,255,255,0.1)"
                  strokeWidth={1}
                />
              )}

              {/* Row background */}
              <rect
                x={margin.left}
                y={0}
                width={dimensions.width - margin.left - margin.right}
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
                  onMouseEnter={(event) => {
                    setTooltip({
                      x: event.clientX,
                      y: event.clientY,
                      activity,
                    });
                  }}
                  onMouseLeave={() => setTooltip(null)}
                />
              ))}

              {/* Bottom row line */}
              <Line
                from={{ x: margin.left, y: ROW_HEIGHT }}
                to={{ x: dimensions.width - margin.right, y: ROW_HEIGHT }}
                stroke="rgba(255,255,255,0.1)"
                strokeWidth={1}
              />
            </Group>
          );
        })}
      </svg>
      {tooltip && (
        <div
          className="absolute z-10 text-white p-2 rounded shadow-lg text-xs custom-tooltip"
          style={{
            left: `${tooltip.x}px`,
            top: `${tooltip.y}px`,
            transform: "translateY(-100%) translateX(-150%)",
            pointerEvents: "none",
          }}
        >
          <div
            className="custom-tooltip-header text-center shadow"
            style={{ color: "white" }}
          >
            {tooltip.activity.project}
          </div>
          <div>
            {format(tooltip.activity.start, "h:mm a")} -{" "}
            {format(tooltip.activity.end, "h:mm a")}
          </div>
          <div>
            Duration:{" "}
            {convertSecondsToHoursAndMinutes(tooltip.activity.duration)}
          </div>
        </div>
      )}
    </div>
  );
};

export default TimeTrackingVisualization;
