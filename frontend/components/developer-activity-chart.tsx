"use client";

import type React from "react";
import { useCallback, useEffect, useRef, useState } from "react";

interface DeveloperActivityChartProps {
  writePercentage?: number;
  readPercentage?: number;
  size?: number;
  strokeWidth?: number;
  className?: string;
  writeColor?: string;
  readColor?: string;
  subText?: string;
}

export default function DeveloperActivityChart({
  writePercentage = 64,
  readPercentage = 36,
  size = 200,
  strokeWidth = 30,
  className = "",
  writeColor = "white",
  readColor = "rgba(255, 255, 255, 0.2)",
  subText,
}: DeveloperActivityChartProps) {
  const [tooltipContent, setTooltipContent] = useState<string | null>(null);
  const [tooltipPosition, setTooltipPosition] = useState({ x: 0, y: 0 });
  const [animatedPercentage, setAnimatedPercentage] = useState(0);
  const chartRef = useRef<HTMLDivElement>(null);

  const radius = size / 2 - strokeWidth / 2;
  const circumference = 2 * Math.PI * radius;
  const writeOffset = circumference * (1 - animatedPercentage / 100);
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

        if (normalizedAngle <= writePercentage * 3.6) {
          setTooltipContent(`Write: 9 hours 45 mins (${writePercentage}%)`);
        } else {
          setTooltipContent(`Read: 5 hours 15 mins (${readPercentage}%)`);
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
      className={`flex items-center justify-center ${className}`}
      ref={chartRef}
    >
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
            strokeDashoffset={writeOffset}
            strokeLinecap="round"
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
            style={{
              transition: "stroke-dashoffset 0.5s ease-out",
            }}
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span
            className="font-bold text-white"
            style={{ fontSize: `${fontSize}px` }}
          >
            {Math.round(animatedPercentage)}%
          </span>
          {subText && (
            <span
              className="text-white mt-2"
              style={{ fontSize: `${subTextSize}px` }}
            >
              {subText}
            </span>
          )}
        </div>
        {tooltipContent && (
          <div
            className="absolute bg-[#1a4971] text-white p-2 rounded shadow-lg text-sm"
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
  );
}
