"use client";

// import { SummariesApiResponse } from "@/lib/types";
import type React from "react";
import { useCallback, useEffect, useRef, useState } from "react";

import { convertSecondsToHoursAndMinutes } from "@/lib/utils";

interface DeveloperActivityChartProps {
  writePercentage?: number;
  readPercentage?: number;
  size?: number;
  strokeWidth?: number;
  className?: string;
  writeColor?: string;
  readColor?: string;
  subText?: string;
  title?: string;
  subtitle?: string;
  totalSeconds: number;
}

export default function DeveloperActivityChart({
  writePercentage = 64,
  size = 170,
  strokeWidth = 20,
  className = "",
  writeColor = "#975ef7", // Red color as seen in screenshot
  readColor = "whitesmoke", // Dark color as seen in screenshot
  subText,
  title = "",
  subtitle,
  totalSeconds,
}: DeveloperActivityChartProps) {
  const [tooltipContent, setTooltipContent] = useState<string | null>(null);
  const [tooltipPosition, setTooltipPosition] = useState({ x: 0, y: 0 });
  const [animatedPercentage, setAnimatedPercentage] = useState(0);
  const chartRef = useRef<HTMLDivElement>(null);

  const readTime = convertSecondsToHoursAndMinutes(
    ((100 - writePercentage) / 100) * totalSeconds
  );
  const writeTime = convertSecondsToHoursAndMinutes(
    (writePercentage / 100) * totalSeconds
  );
  const readPercentage = 100 - writePercentage;

  const radius = size / 2 - strokeWidth / 2;
  const circumference = 2 * Math.PI * radius;
  const readOffset = circumference * (1 - animatedPercentage / 100);
  const fontSize = size / 4.8;
  const subTextSize = size / 16;
  const center = size / 2;

  useEffect(() => {
    const animationDuration = 1000; // 1 second
    const steps = 60; // 60 frames per second
    const increment = writePercentage / steps;
    let currentPercentage = 0;

    const intervalId = setInterval(() => {
      currentPercentage += increment;
      if (currentPercentage >= writePercentage) {
        clearInterval(intervalId);
        setAnimatedPercentage(writePercentage);
      } else {
        setAnimatedPercentage(currentPercentage);
      }
    }, animationDuration / steps);

    return () => clearInterval(intervalId);
  }, [writePercentage]);

  const handleMouseMove = useCallback(
    (e: React.MouseEvent<SVGCircleElement>) => {
      if (chartRef.current) {
        const rect = chartRef.current.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        const dx = x - center;
        const dy = y - center;
        const angle = Math.atan2(dy, dx) * (180 / Math.PI);
        const normalizedAngle = (angle + 90 + 360) % 360;

        if (normalizedAngle <= readPercentage * 3.6) {
          setTooltipContent(`Read: ${Math.round(readPercentage)}%`);
        } else {
          setTooltipContent(`Write: ${Math.round(writePercentage)}%`);
        }

        setTooltipPosition({
          x: e.clientX - rect.left,
          y: e.clientY - rect.top,
        });
      }
    },
    [writePercentage, readPercentage, center]
  );

  const handleMouseLeave = useCallback(() => {
    setTooltipContent(null);
  }, []);

  return (
    <div
      className={`flex flex-col items-center justify-center ${className}`}
      style={{ height: "190px", width: "100%" }}
    >
      {/* Title at the top */}
      <h4 className="text-xs font-medium mb-0">
        {title} {""}
      </h4>
      {subtitle && (
        <span className="text-xs text-gray-500 mb-1">{subtitle}</span>
      )}

      <div className="flex flex-col items-center justify-center" ref={chartRef}>
        <div className="relative" style={{ width: size, height: size }}>
          <svg
            width={size}
            height={size}
            viewBox={`0 0 ${size} ${size}`}
            className="transform -rotate-90"
          >
            <circle
              cx={center}
              cy={center}
              r={radius}
              fill="none"
              stroke={readColor}
              strokeWidth={strokeWidth}
              onMouseMove={handleMouseMove}
              onMouseLeave={handleMouseLeave}
            />
            <circle
              cx={center}
              cy={center}
              r={radius}
              fill="none"
              stroke={writeColor}
              strokeWidth={strokeWidth}
              strokeDasharray={circumference}
              strokeDashoffset={readOffset}
              strokeLinecap="round"
              onMouseMove={handleMouseMove}
              onMouseLeave={handleMouseLeave}
              style={{
                transition: "stroke-dashoffset 0.5s ease-out",
              }}
            />
          </svg>
          <div className="absolute inset-0 flex flex-col items-center justify-center">
            <span className="text-xs">Write Time</span>
            <span className="font-bold" style={{ fontSize: `${fontSize}px` }}>
              {Math.round(animatedPercentage)}%
            </span>
            {subText && (
              <span
                className="text-xs mt-1"
                style={{ fontSize: `${subTextSize}px` }}
              >
                {subText}
              </span>
            )}
          </div>
          {tooltipContent && (
            <div
              className="absolute bg-[#1a4971] text-white p-1 rounded shadow-lg text-xs"
              style={{
                left: `${tooltipPosition.x}px`,
                top: `${tooltipPosition.y}px`,
                transform: "translate(-50%, -100%)",
                pointerEvents: "none",
              }}
            >
              {tooltipContent}
            </div>
          )}
        </div>
      </div>

      {/* Legend */}
      <div className="flex justify-between gap-2 mt-2 w-full px-4">
        <div className="flex items-center gap-1">
          <div
            className="w-3 h-3 rounded-full"
            style={{ backgroundColor: readColor }}
          ></div>
          <span className="text-xs">Read: {readTime}</span>
        </div>
        <div className="flex items-center gap-1">
          <div
            className="w-3 h-3 rounded-full"
            style={{ backgroundColor: writeColor }}
          ></div>
          <span className="text-xs">Write: {writeTime}</span>
        </div>
      </div>
    </div>
  );
}
