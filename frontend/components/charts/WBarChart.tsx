"use client";

import { AxisBottom, AxisLeft } from "@visx/axis";
import { Group } from "@visx/group";
import { LegendOrdinal } from "@visx/legend";
import { scaleBand, scaleLinear } from "@visx/scale";
import { scaleOrdinal } from "@visx/scale";
import { Bar } from "@visx/shape";
import { useTooltip, useTooltipInPortal, defaultStyles } from "@visx/tooltip";
import { ScaleBand, ScaleLinear } from "d3-scale";
import { BarChart2, BarChartHorizontal } from "lucide-react";
import React, { useEffect, useMemo, useRef, useState } from "react";

import { useMediaQuery } from "@/hooks/use-media-query";
import {
  convertSecondsToHoursAndMinutes,
  getEntityColor,
  getMachineColor,
} from "@/lib/utils";

import { Button } from "../ui/button";
import { EmptyChartWrapper } from "./EmptyChartWrapper";

export interface WBarChartDataItem {
  key: string;
  total: number;
  percent?: number;
  formattedDuration?: string;
  formattedPercent?: string;
}

export interface WBarChartProps {
  data: WBarChartDataItem[];
  title: string;
  innerRadius?: number; // Kept for compatibility, unused
  colorNamespace: string;
  durationSubtitle?: string;
  defaultOrientation?: "vertical" | "horizontal"; // New prop for orientation
}

interface Dimensions {
  width: number;
  height: number;
}

const useComponentDimensions = (
  ref: React.RefObject<HTMLDivElement>
): Dimensions => {
  const [dimensions, setDimensions] = useState<Dimensions>({
    width: 0,
    height: 0,
  });

  useEffect(() => {
    if (ref.current) {
      const resizeObserver = new ResizeObserver((entries) => {
        const { width, height } = entries[0].contentRect;
        setDimensions({ width, height });
      });
      resizeObserver.observe(ref.current);
      return () => resizeObserver.disconnect();
    }
  }, [ref]);

  return dimensions;
};

// Enhanced tooltip styles
const tooltipStyles = {
  ...defaultStyles,
  backgroundColor: "white",
  color: "#333",
  padding: "8px 12px",
  border: "1px solid #ccc",
  borderRadius: "4px",
  boxShadow: "0 2px 10px rgba(0,0,0,0.1)",
  fontSize: "12px",
  zIndex: 1000,
};

function WBarChartComponent({
  data,
  title,
  colorNamespace,
  defaultOrientation = "vertical",
}: WBarChartProps) {
  const [orientation, setOrientation] = useState<"vertical" | "horizontal">(
    defaultOrientation
  );
  const containerRef = useRef<HTMLDivElement>(null);
  const { width: containerWidth, height: containerHeight } =
    useComponentDimensions(containerRef);
  const width = containerWidth || 600;
  const height = containerHeight || 200;

  const isMobile = useMediaQuery("only screen and (max-width : 576px)");

  // Hide legend when in horizontal mode
  const hideLegend = isMobile || orientation === "horizontal";

  const bottom = isMobile && orientation === "vertical" ? 60 : 30;

  let left = 0;
  if (orientation === "horizontal") {
    left = 60;
  }

  const margin = { top: 5, right: 20, bottom, left }; // Increased left margin for horizontal labels

  // Calculate percentage for each item
  const totalValue = useMemo(
    () => data.reduce((sum, item) => sum + item.total, 0),
    [data]
  );

  const dataWithPercent = useMemo(
    () =>
      data.map((item) => ({
        ...item,
        percent: totalValue > 0 ? item.total / totalValue : 0,
        formattedDuration: convertSecondsToHoursAndMinutes(item.total),
        formattedPercent:
          totalValue > 0
            ? `${((item.total / totalValue) * 100).toFixed(0)}%`
            : "0%",
      })),
    [data, totalValue]
  );

  // Sort data by total (descending)
  const sortedData = useMemo(
    () => [...dataWithPercent].sort((a, b) => b.total - a.total),
    [dataWithPercent]
  );

  // Tooltip setup with enhanced configuration
  const {
    tooltipData,
    tooltipLeft,
    tooltipTop,
    tooltipOpen,
    showTooltip,
    hideTooltip,
  } = useTooltip<WBarChartDataItem>({
    // Set initial state to ensure tooltip is active
    tooltipOpen: false,
    tooltipLeft: 0,
    tooltipTop: 0,
  });

  // Create a parent reference for the tooltip portal
  const parentRef = useRef<HTMLDivElement>(null);

  // Use the parent element as the container for the tooltip portal
  const { containerRef: tooltipContainerRef, TooltipInPortal } =
    useTooltipInPortal({
      scroll: true,
      detectBounds: true,
      // This ensures the tooltip is positioned relative to the chart container
      // parentRef: parentRef as any,
    });

  // Format seconds to hours and minutes for axis ticks
  const formatSeconds = (seconds: number): string => {
    return convertSecondsToHoursAndMinutes(seconds);
  };

  // Scales
  const xScale = useMemo(() => {
    if (orientation === "vertical") {
      return scaleBand<string>({
        domain: sortedData.map((d) => d.key),
        range: [0, width - margin.left - margin.right],
        padding: 0.3,
      });
    } else {
      return scaleLinear<number>({
        domain: [0, Math.max(...sortedData.map((d) => d.total))],
        range: [0, width - margin.left - margin.right],
        nice: true,
      });
    }
  }, [sortedData, width, orientation, margin.left, margin.right]) as
    | ScaleBand<string>
    | ScaleLinear<number, number>;

  const yScale = useMemo(() => {
    if (orientation === "vertical") {
      return scaleLinear<number>({
        domain: [0, Math.max(...sortedData.map((d) => d.total))],
        range: [height - margin.top - margin.bottom, 0],
        nice: true,
      });
    } else {
      return scaleBand<string>({
        domain: sortedData.map((d) => d.key),
        range: [0, height - margin.top - margin.bottom],
        padding: 0.3,
      });
    }
  }, [sortedData, height, orientation, margin.bottom, margin.top]) as
    | ScaleLinear<number, number>
    | ScaleBand<string>;

  // Color scale for legend
  const colorScale = scaleOrdinal<string, string>({
    domain: sortedData.map((d) => d.key),
    range: sortedData.map((d) =>
      colorNamespace !== "machines"
        ? getEntityColor(colorNamespace, d.key)
        : getMachineColor(d.key)
    ),
  });

  // Completely revised tooltip handling function
  const handleMouseOver = (
    event: React.MouseEvent<SVGRectElement>,
    data: WBarChartDataItem
  ) => {
    // Get the SVG element's bounding rect
    const svgBounds =
      event.currentTarget.ownerSVGElement?.getBoundingClientRect();
    // Get the bar element's bounding rect
    const barBounds = event.currentTarget.getBoundingClientRect();

    if (!svgBounds) return;

    // Calculate the position relative to the SVG
    const svgX = barBounds.left - svgBounds.left + barBounds.width / 2;
    const svgY = barBounds.top - svgBounds.top;

    // Position tooltip with offsets
    let tooltipX = svgX;
    let tooltipY = svgY - 10; // Position above the bar

    // Adjust for orientation
    if (orientation === "horizontal") {
      tooltipX = barBounds.right - svgBounds.left + 10; // Position to the right of the bar
      tooltipY = barBounds.top - svgBounds.top + barBounds.height / 2;
    }

    // Debug to help with positioning issues
    console.log("Tooltip position:", {
      tooltipX,
      tooltipY,
      barBounds,
      svgBounds,
      orientation,
    });

    showTooltip({
      tooltipData: data,
      tooltipLeft: tooltipX,
      tooltipTop: tooltipY,
    });
  };

  if (!width || !height) {
    return (
      <div ref={containerRef} style={{ width: "100%", height: "300px" }}></div>
    );
  }

  const isHorizontal = orientation === "horizontal";

  return (
    // Use parentRef to create a context for the tooltip positioning
    <div ref={parentRef} style={{ position: "relative", width: "100%" }}>
      <div className="chart-box-title">
        {title}
        <Button
          variant="link"
          size="icon"
          onClick={() => {
            setOrientation(isHorizontal ? "vertical" : "horizontal");
          }}
          className="p-0 transition"
        >
          {isHorizontal ? (
            <BarChart2 size={20} />
          ) : (
            <BarChartHorizontal size={20} />
          )}
        </Button>
      </div>
      <div
        ref={containerRef}
        style={{
          width: "100%",
          height: "200px",
          position: "relative",
        }}
      >
        <svg width={width} height={height} ref={tooltipContainerRef}>
          <Group left={margin.left} top={margin.top}>
            {orientation === "vertical" ? (
              // Vertical bars
              <>
                {/* Only show X-axis labels on mobile */}
                {isMobile && (
                  <AxisBottom
                    top={height - margin.top - margin.bottom}
                    scale={xScale as ScaleBand<string>}
                    stroke="#ccc"
                    hideAxisLine
                    hideTicks
                    tickLabelProps={() => ({
                      fill: "#888",
                      fontSize: 10,
                      textAnchor: "middle",
                      dy: 10,
                    })}
                    tickComponent={({ x, y, formattedValue }) => (
                      <text
                        x={x}
                        y={y}
                        dy={10}
                        textAnchor="end"
                        transform={`rotate(-45, ${x}, ${y})`}
                        fontSize={10}
                        fill="#888"
                      >
                        {formattedValue}
                      </text>
                    )}
                  />
                )}

                {sortedData.map((d, i) => {
                  const barWidth = (xScale as ScaleBand<string>).bandwidth();
                  const barHeight =
                    height -
                    margin.top -
                    margin.bottom -
                    (yScale as ScaleLinear<number, number>)(d.total);
                  const barX = xScale(d.key as any);
                  const barY = (yScale as ScaleLinear<number, number>)(d.total);
                  const fill =
                    colorNamespace !== "machines"
                      ? getEntityColor(colorNamespace, d.key)
                      : getMachineColor(d.key);

                  return (
                    <React.Fragment key={`bar-${i}`}>
                      <Bar
                        x={barX}
                        y={barY}
                        width={barWidth}
                        height={barHeight}
                        fill={fill}
                        stroke="black"
                        strokeWidth={1}
                        onMouseMove={(event) => handleMouseOver(event, d)}
                        onMouseLeave={hideTooltip}
                        style={{
                          cursor: "pointer",
                          transition: "all 0.3s ease",
                        }}
                      />
                    </React.Fragment>
                  );
                })}
              </>
            ) : (
              // Horizontal bars
              <>
                <AxisBottom
                  top={height - margin.top - margin.bottom}
                  scale={xScale as ScaleLinear<number, number>}
                  stroke="#ccc"
                  hideAxisLine
                  hideTicks
                  numTicks={5}
                  tickFormat={formatSeconds as any}
                  tickLabelProps={() => ({
                    fill: "#888",
                    fontSize: 10,
                    textAnchor: "middle",
                  })}
                />

                <AxisLeft
                  scale={yScale as ScaleBand<string>}
                  stroke="#ccc"
                  hideAxisLine
                  hideTicks
                  tickLabelProps={() => ({
                    fill: "#888",
                    fontSize: 10,
                    textAnchor: "end",
                    dy: "0.33em",
                    dx: -5,
                  })}
                />

                {sortedData.map((d, i) => {
                  const barHeight = (yScale as ScaleBand<string>).bandwidth();
                  const barWidth = (xScale as ScaleLinear<number, number>)(
                    d.total
                  );
                  const barX = 0;
                  const barY = yScale(d.key as any);
                  const fill =
                    colorNamespace !== "machines"
                      ? getEntityColor(colorNamespace, d.key)
                      : getMachineColor(d.key);

                  return (
                    <React.Fragment key={`bar-${i}`}>
                      <Bar
                        x={barX}
                        y={barY}
                        width={barWidth}
                        height={barHeight}
                        fill={fill}
                        stroke="grey"
                        strokeWidth={1}
                        onMouseMove={(event) => handleMouseOver(event, d)}
                        onMouseLeave={hideTooltip}
                        style={{
                          cursor: "pointer",
                          transition: "all 0.3s ease",
                        }}
                      />
                    </React.Fragment>
                  );
                })}
              </>
            )}
          </Group>
        </svg>

        {!hideLegend && (
          <div
            style={{
              position: "absolute",
              top: 10,
              right: 20,
              fontSize: "9px",
              maxHeight: "70%",
              overflowY: "auto",
            }}
          >
            <LegendOrdinal
              scale={colorScale}
              direction="column"
              labelMargin="0 15px 0 0"
              shape="circle"
              style={{ fontSize: "9px" }}
              itemDirection="row"
              shapeHeight={10}
              shapeWidth={10}
              labelFormat={(label) => {
                const item = sortedData.find((d) => d.key === label);
                return `${label} - ${item?.formattedDuration} (${item?.formattedPercent})`;
              }}
            />
          </div>
        )}

        {/* Enhanced tooltip with improved positioning */}
        {tooltipOpen && tooltipData && (
          <TooltipInPortal
            top={tooltipTop}
            left={tooltipLeft}
            style={tooltipStyles}
            // Add applyPositionStyle to ensure proper positioning
            applyPositionStyle={true}
          >
            <div style={{ fontWeight: "bold" }}>{tooltipData.key}</div>
            <div>
              {tooltipData.formattedDuration} ({tooltipData.formattedPercent})
            </div>
          </TooltipInPortal>
        )}
      </div>
    </div>
  );
}

export function WBarChart({
  data,
  title,
  innerRadius,
  colorNamespace,
  durationSubtitle,
  defaultOrientation = "vertical",
}: WBarChartProps) {
  return (
    <EmptyChartWrapper hasData={Object.keys(data).length > 0}>
      <WBarChartComponent
        data={data}
        title={title}
        innerRadius={innerRadius}
        colorNamespace={colorNamespace}
        durationSubtitle={durationSubtitle}
        defaultOrientation={defaultOrientation}
      />
    </EmptyChartWrapper>
  );
}
