"use client";

import { Group } from "@visx/group";
import { Pack } from "@visx/hierarchy";
import { ParentSize } from "@visx/responsive";
import { hierarchy } from "d3-hierarchy";
import truncate from "lodash/truncate";
import React, { useMemo } from "react";

// Color palette for bubbles
const colors = [
  "#8ed3c7",
  "#fb8072",
  "#ffffb2",
  "#bebada",
  "#b3de68",
  "#b3de68",
  "#ffed6f",
];

interface FileData {
  name: string;
  digital: string;
  total_seconds: number;
  text: string;
}

interface ProcessedFileData {
  name: string;
  parentFolder: string;
  size: number;
  timeText: string;
  fullPath: string;
}

interface HierarchyNode {
  name: string;
  children?: ProcessedFileData[];
}

interface BubbleChartProps {
  rawData: FileData[];
}

const FileActivityBubble: React.FC<BubbleChartProps> = ({ rawData }) => {
  console.log("rawData:", rawData.slice(0, 5));
  const data = useMemo(() => {
    const topFiles = rawData
      .filter((d): d is FileData => d.total_seconds > 0)
      .sort((a, b) => b.total_seconds - a.total_seconds)
      .slice(0, 20) // Reduced to top 20 files
      .map((d) => {
        const pathParts = d.name.split("/");
        const fileName = pathParts.pop() || "";
        const parentFolder = pathParts.pop() || "";

        // Format time text to always show "Xhrs Ymins"
        const totalSeconds = d.total_seconds;
        const hours = Math.floor(totalSeconds / 3600);
        const minutes = Math.floor((totalSeconds % 3600) / 60);

        const hourTimeText = hours > 0 ? `${hours}hrs` : "";
        const minuteTimeText = minutes > 0 ? `${minutes}mins` : "";
        const spacing = hourTimeText && minuteTimeText ? " " : "";
        const timeText = `${hourTimeText}${spacing}${minuteTimeText}`;

        return {
          name: fileName,
          parentFolder,
          size: d.total_seconds,
          timeText,
          fullPath: d.name,
        };
      });

    return {
      name: "root",
      children: topFiles,
    } as HierarchyNode;
  }, [rawData]);

  const getColor = (index: number): string => {
    return colors[index % colors.length];
  };

  // Make font size relative to the bubble size
  const getFontSize = (radius: number): number => {
    return Math.max(8, Math.min(14, radius * 0.3)); // Adjust scaling factor as needed
  };

  const truncateFolderName = (
    folder: string,
    fileName: string,
    maxLength: number
  ): string => {
    // If the full path fits within maxLength, return it as is
    const combinedText = `${folder}/${fileName}`;
    if (combinedText.length <= maxLength) return combinedText;

    // If the filename itself is longer than maxLength, truncate the filename
    if (fileName.length >= maxLength) {
      return truncate(fileName, {
        length: maxLength,
        omission: "...",
      });
    }

    // Calculate remaining space for the folder after including the filename and "/"
    const remainingLength = maxLength - fileName.length - 1; // Subtract 1 for the "/"

    // If there's no space for the folder, return only the filename
    if (remainingLength <= 0) return fileName;

    // Truncate the folder to fit the remaining space
    const truncatedFolder = truncate(folder, {
      length: remainingLength,
      omission: "...",
    });

    return `${truncatedFolder}/${fileName}`;
  };

  return (
    <div className="relative">
      <ParentSize>
        {({ width }) => {
          const height = Math.max(600, width * 0.75);
          return (
            <svg width={width} height={height}>
              <rect width={width} height={height} fill="none" rx={14} />
              <Pack<HierarchyNode>
                root={hierarchy(data)
                  .sum((d) => (d as unknown as ProcessedFileData)?.size || 0)
                  .sort((a, b) => (b.value || 0) - (a.value || 0))}
                size={[width - 16, height - 16]}
                padding={16}
              >
                {(packData) => {
                  const circles = packData.descendants().slice(1);
                  return (
                    <Group top={8} left={8}>
                      {circles.map((circle, i) => {
                        const node =
                          circle.data as unknown as ProcessedFileData;
                        const fontSize = getFontSize(circle.r);
                        const maxTextLength = Math.floor(
                          (circle.r * 2) / (fontSize * 0.6)
                        ); // Adjust based on font size

                        const displayText = truncateFolderName(
                          node.parentFolder,
                          node.name,
                          maxTextLength
                        );

                        return (
                          <Group key={`circle-${i}`}>
                            <circle
                              r={circle.r}
                              cx={circle.x}
                              cy={circle.y}
                              fill={getColor(i)}
                              opacity={0.9}
                              className="transition-transform duration-150 hover:opacity-80"
                              transform-origin={`${circle.x} ${circle.y}`}
                              onMouseEnter={(e) => {
                                const target = e.target as SVGCircleElement;
                                target.setAttribute("transform", "scale(1.02)");
                              }}
                              onMouseLeave={(e) => {
                                const target = e.target as SVGCircleElement;
                                target.setAttribute("transform", "scale(1)");
                              }}
                            >
                              <title>{`${node.fullPath}\n${node.timeText}`}</title>
                            </circle>
                            {fontSize > 0 && (
                              <text
                                x={circle.x}
                                y={circle.y}
                                fontSize={fontSize}
                                fill="#000"
                                textAnchor="middle"
                                className="font-medium pointer-events-none"
                              >
                                <tspan x={circle.x} dy="0">
                                  {displayText}
                                </tspan>
                                <tspan
                                  x={circle.x}
                                  dy="1.2em"
                                  fontSize={fontSize * 0.9} // Slightly smaller font for time
                                >
                                  <tspan fill="#000">{node.timeText}</tspan>
                                </tspan>
                              </text>
                            )}
                          </Group>
                        );
                      })}
                    </Group>
                  );
                }}
              </Pack>
            </svg>
          );
        }}
      </ParentSize>
    </div>
  );
};

export default FileActivityBubble;
