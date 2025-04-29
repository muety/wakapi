"use client";

import { Group } from "@visx/group";
import { scaleLinear, scaleTime } from "@visx/scale";
import { Bar, Line } from "@visx/shape";
import { addDays, format, isToday, subDays } from "date-fns";
import { startCase } from "lodash";
import { ChevronLeft, ChevronRight, FileBarChart } from "lucide-react";
import { useRouter } from "next/navigation";
import type React from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { COLORS } from "@/lib/constants";
import { convertSecondsToHoursAndMinutes } from "@/lib/utils";

interface RawTimeEntry {
  time: number;
  project: string;
  duration: number;
  color: string | null;
  [key: string]: any; // Allow for arbitrary keys based on sliceBy
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
  project: string; // Keep project for potential display
  duration: number;
  [key: string]: any; // Allow for arbitrary keys based on sliceBy
}

interface DataGroup {
  name: string; // This will be the value of the sliceBy key
  activities: ProcessedActivity[];
  totalDuration: number;
}

interface TimeTrackingProps {
  data: RawData;
  margin?: { top: number; right: number; bottom: number; left: number };
  sliceBy?: string; // New prop to specify the key for grouping
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

// Empty state component
function EmptyState({ date }: { date: Date }) {
  const formattedDate = date.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  return (
    <div className="flex flex-col items-center justify-center h-96 w-full">
      <div className="bg-blue-500/10 rounded-full p-6 mb-6">
        <FileBarChart className="size-16 text-blue-500" />
      </div>
      <h3 className="text-2xl font-bold text-white mb-2">
        No time entries found
      </h3>
      <p className="text-gray-400 mb-6 text-center max-w-md">
        There are no time entries recorded for {formattedDate}.
      </p>
    </div>
  );
}

interface BarModalProps {
  activity: ProcessedActivity;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  sliceBy?: string;
}

const BarModal: React.FC<BarModalProps> = ({
  activity,
  open,
  onOpenChange,
  sliceBy = "project",
}) => {
  const durationString = convertSecondsToHoursAndMinutes(activity.duration);
  const title = startCase(activity[sliceBy] || "Unknown");
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-zinc-900 text-zinc-100 border-zinc-800">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>Details of this activity.</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 gap-2">
            <div className="text-zinc-400">{startCase(sliceBy)}:</div>
            <div className="col-span-3 font-semibold">{title}</div>
          </div>
          <div className="grid grid-cols-4 gap-2">
            <div className="text-zinc-400">Duration:</div>
            <div className="col-span-3">{durationString}</div>
          </div>
          <div className="grid grid-cols-4 gap-2">
            <div className="text-zinc-400">Start Time:</div>
            <div className="col-span-3">
              {format(activity.start, "yyyy-MM-dd hh:mm a")}
            </div>
          </div>
          <div className="grid grid-cols-4 gap-2">
            <div className="text-zinc-400">End Time:</div>
            <div className="col-span-3">
              {format(activity.end, "yyyy-MM-dd hh:mm a")}
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

const TimeTrackingVisualization: React.FC<TimeTrackingProps> = ({
  data: rawData,
  margin = defaultMargin,
  sliceBy = "project", // Default to 'project'
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({ width: 800, height: 600 });
  const [modalActivity, setModalActivity] = useState<ProcessedActivity | null>(
    null
  );
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [tooltip, setTooltip] = useState<{
    x: number;
    y: number;
    activity: ProcessedActivity;
  } | null>(null);

  const handleBarClick = (activity: ProcessedActivity) => {
    setModalActivity(activity);
    setIsModalOpen(true);
  };

  const getColor = useCallback(
    (activity: ProcessedActivity) => {
      const defaultColor = "#3b82f6"; // Default color if no specific color is found
      if (sliceBy === "language") {
        return COLORS.languages[activity.language] || defaultColor;
      }
      if (sliceBy === "editor") {
        return COLORS.editors[activity.editor] || defaultColor;
      }

      if (sliceBy === "category") {
        return COLORS.categories[activity.category] || defaultColor;
      }
      return defaultColor;
    },
    [sliceBy]
  );

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

  const { timeScale, yScale, groups, totalTime, calculatedHeight } =
    useMemo(() => {
      const startDate = new Date(rawData.start);
      const endDate = new Date(rawData.end);

      const activities: ProcessedActivity[] = rawData.data.map((item) => ({
        id: item.time.toString(),
        start: new Date(item.time * 1000),
        end: new Date((item.time + item.duration) * 1000),
        project: item.project, // Keep project data
        duration: item.duration,
        ...Object.fromEntries(
          Object.keys(item)
            .filter((key) => key !== "time" && key !== "duration")
            .map((key) => [key, item[key]])
        ), // Include other potential slicing keys
      }));

      const dataGroups: { [key: string]: ProcessedActivity[] } = {};
      activities.forEach((activity) => {
        const groupKey = activity[sliceBy] || "Unknown";
        if (!dataGroups[groupKey]) {
          dataGroups[groupKey] = [];
        }
        dataGroups[groupKey].push(activity);
      });

      const groups: DataGroup[] = Object.entries(dataGroups)
        .map(([name, activities]) => ({
          name,
          activities,
          totalDuration: activities.reduce((sum, act) => sum + act.duration, 0),
        }))
        .sort((a, b) => b.totalDuration - a.totalDuration); // Sort by total duration

      const totalTime = groups.reduce(
        (sum, group) => sum + group.totalDuration,
        0
      );

      // Calculate height based on number of groups
      const calculatedHeight =
        groups.length * ROW_HEIGHT + margin.top + margin.bottom;

      const timeScale = scaleTime({
        domain: [startDate, endDate],
        range: [margin.left, dimensions.width - margin.right],
      });

      const yScale = scaleLinear({
        domain: [0, groups.length],
        range: [margin.top, calculatedHeight - margin.bottom],
      });

      return { timeScale, yScale, groups, totalTime, calculatedHeight };
    }, [dimensions.width, margin, rawData, sliceBy]); // Add sliceBy to dependency array

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

  // Check if there's no data
  const hasNoData = rawData.data.length === 0;

  return (
    <div
      ref={containerRef}
      className="relative bg-[rgb(18,18,18)] rounded-xl p-4 w-full h-full"
    >
      <DayHeader data={rawData} totalTime={totalTime} />

      {hasNoData ? (
        <EmptyState date={new Date(rawData.start)} />
      ) : (
        <>
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

            {/* Time axis with ticks below labels */}
            <Group top={margin.top - 25}>
              {timeScale.ticks(tickCount).map((date, i) => {
                const x = timeScale(date);
                const label = getHourLabel(date);
                const isEdge = i === 0 || i === tickCount - 1;

                return (
                  <g key={i}>
                    <text
                      x={x}
                      y={15}
                      textAnchor={
                        isEdge ? (i === 0 ? "start" : "end") : "middle"
                      }
                      fill="#6b7280"
                      fontSize={12}
                      fontFamily="monospace"
                    >
                      {label}
                    </text>
                    <Line
                      from={{ x, y: 20 }}
                      to={{ x, y: 25 }}
                      stroke="rgba(255,255,255,0.1)"
                      strokeWidth={1}
                    />
                  </g>
                );
              })}
            </Group>

            {/* Data Group lanes */}
            {groups.map((group, index) => {
              const yPos = yScale(index);
              return (
                <Group key={group.name} top={yPos}>
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

                  {/* Group label */}
                  <text
                    x={margin.left - 20}
                    y={ROW_HEIGHT / 2 + 10}
                    fill="#fff"
                    textAnchor="end"
                    fontSize={14}
                    fontFamily="monospace"
                  >
                    {`${group.name} ${formatDuration(group.totalDuration)}`}
                  </text>

                  {/* Activity bars */}
                  {group.activities.map((activity) => (
                    <Bar
                      key={activity.id}
                      x={timeScale(activity.start)}
                      y={(ROW_HEIGHT - LANE_HEIGHT) / 2}
                      width={Math.max(
                        2,
                        timeScale(activity.end) - timeScale(activity.start)
                      )}
                      height={LANE_HEIGHT}
                      fill={getColor(activity)} // Consider using a color based on the grouping if available
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
                      onClick={() => handleBarClick(activity)}
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
                {/* Display the value of the sliceBy key for the tooltip header */}
                {tooltip.activity[sliceBy] || "Unknown"}
              </div>
              <div>
                {format(tooltip.activity.start, "h:mm a")} -{" "}
                {format(tooltip.activity.end, "h:mm a")}
              </div>
              <div>
                Duration:{" "}
                {convertSecondsToHoursAndMinutes(tooltip.activity.duration)}
              </div>
              {/* Optionally display the project name as well */}
              {sliceBy !== "project" && (
                <div>Project: {tooltip.activity.project}</div>
              )}
            </div>
          )}
        </>
      )}

      {modalActivity && (
        <BarModal
          activity={modalActivity}
          open={isModalOpen}
          onOpenChange={setIsModalOpen}
          sliceBy={sliceBy}
        />
      )}
    </div>
  );
};

export default TimeTrackingVisualization;
