"use client";

import type React from "react";
import { useMemo, useState } from "react";
import { Group } from "@visx/group";
import { Treemap } from "@visx/hierarchy";
import { ParentSize } from "@visx/responsive";
import { hierarchy } from "d3-hierarchy";
import { scaleOrdinal } from "@visx/scale";
import truncate from "lodash/truncate";

// Color palette
const colors = [
  "#8ed3c7",
  "#fb8072",
  "#ffffb2",
  "#bebada",
  "#b3de68",
  "#ffed6f",
];

interface FileData {
  name: string;
  digital: string;
  total_seconds: number;
  text: string;
}

interface TreemapChartProps {
  rawData: FileData[];
}

export const FileActivityTreemapVisx: React.FC<TreemapChartProps> = ({
  rawData,
}) => {
  // State for hover and selection interactions
  const [hoveredNode, setHoveredNode] = useState<any | null>(null);
  const [selectedFolder, setSelectedFolder] = useState<string | null>(null);

  const data = useMemo(() => {
    // Process the raw data and deduplicate files
    const fileMap = new Map<string, FileData>();

    // First pass: keep only the highest total_seconds for each filename
    rawData.forEach((file) => {
      if (file.total_seconds <= 0) return;

      const existingFile = fileMap.get(file.name);
      if (!existingFile || file.total_seconds > existingFile.total_seconds) {
        fileMap.set(file.name, file);
      }
    });

    // Convert map to array and sort
    const topFiles = Array.from(fileMap.values())
      .sort((a, b) => b.total_seconds - a.total_seconds)
      .slice(0, 20); // Top 20 files

    // Create a hierarchical structure for the treemap
    interface HierarchyNode {
      name: string;
      children?: HierarchyNode[];
      size?: number;
      fullPath?: string;
      timeText?: string;
      id?: string;
    }

    const root: HierarchyNode = { name: "root", children: [] };
    const folderMap: Record<string, any> = {};

    topFiles.forEach((file) => {
      // Split the path into parts
      const pathParts = file.name.split("/").filter(Boolean);
      const fileName = pathParts.pop() || "";

      // Format time text
      const totalSeconds = file.total_seconds;
      const hours = Math.floor(totalSeconds / 3600);
      const minutes = Math.floor((totalSeconds % 3600) / 60);
      const seconds = Math.floor(totalSeconds % 60);

      let timeText = "";
      if (hours > 0) timeText += `${hours}hrs `;
      if (minutes > 0) timeText += `${minutes}mins `;
      if (seconds > 0 && hours === 0 && minutes === 0)
        timeText += `${seconds}s`;
      timeText = timeText.trim();

      // Get parent folder
      const parentFolder =
        pathParts.length > 0 ? pathParts[pathParts.length - 1] : "root";

      // Create folder id to ensure uniqueness
      const folderId = pathParts.join("/");

      // Create folder if it doesn't exist
      if (!folderMap[folderId]) {
        const newFolder = {
          name: parentFolder,
          children: [],
          id: folderId,
        };
        folderMap[folderId] = newFolder;
        root.children = root.children || [];
        root.children.push(newFolder);
      }

      // Add file to folder
      folderMap[folderId].children.push({
        name: fileName,
        size: file.total_seconds,
        fullPath: file.name,
        timeText,
        id: file.name,
      });
    });

    return root;
  }, [rawData]);

  // Create a color scale
  const colorScale = scaleOrdinal({
    domain: colors.map((_, i) => `${i}`),
    range: colors,
  });

  return (
    <div className="relative">
      {/* Custom tooltip that follows mouse */}
      {hoveredNode && (
        <div
          className="absolute bg-white p-2 rounded shadow-lg border border-gray-200 z-10 max-w-xs"
          style={{
            left: `${Math.min(window.innerWidth - 250, hoveredNode.x + 10)}px`,
            top: `${hoveredNode.y + 10}px`,
          }}
        >
          <p className="font-bold text-sm mb-1">{hoveredNode.name}</p>
          {hoveredNode.fullPath && (
            <p className="text-xs text-gray-600 mb-1">{hoveredNode.fullPath}</p>
          )}
          {hoveredNode.timeText && (
            <p className="text-xs font-medium">{hoveredNode.timeText}</p>
          )}
        </div>
      )}

      {/* Folder filters */}
      <div className="mb-4 flex flex-wrap gap-2">
        <button
          className={`px-3 py-1 text-sm rounded ${selectedFolder === null ? "bg-blue-500 text-white" : ""}`}
          onClick={() => setSelectedFolder(null)}
        >
          All Folders
        </button>
        {data.children?.map((folder) => (
          <button
            key={folder.id}
            className={`px-3 py-1 text-sm rounded ${selectedFolder === folder.id ? "bg-blue-500 text-white" : ""}`}
            onClick={() =>
              setSelectedFolder(
                folder.id === selectedFolder ? null : folder.id || ""
              )
            }
          >
            {folder.name}
          </button>
        ))}
      </div>

      <ParentSize>
        {({ width }) => {
          const height = 600;

          // Filter data based on selection
          const displayData = useMemo(() => {
            if (!selectedFolder) return data;

            return {
              name: "root",
              children:
                data.children?.filter(
                  (folder) => folder.id === selectedFolder
                ) || [],
            };
          }, [data, selectedFolder]);

          return (
            <svg width={width} height={height}>
              <Treemap<typeof displayData>
                root={hierarchy(displayData)
                  .sum((d) => (d as any).size || 0)
                  .sort((a, b) => (b.value || 0) - (a.value || 0))}
                size={[width, height]}
                padding={3}
                round
              >
                {(treemap) => (
                  <Group>
                    {treemap
                      .descendants()
                      .filter((node) => node.depth > 0) // Skip root
                      .map((node, i) => {
                        const nodeWidth = node.x1 - node.x0;
                        const nodeHeight = node.y1 - node.y0;
                        const isFolder = node.depth === 1;
                        const fontSize = Math.min(
                          14,
                          Math.max(10, nodeWidth / 10)
                        );

                        // Skip rendering text if rectangle is too small
                        const showText = nodeWidth > 30 && nodeHeight > 30;

                        // Check if node is hovered
                        const isHovered = hoveredNode?.id === node.data.id;

                        return (
                          <Group
                            key={`node-${node.data.id || i}`}
                            onMouseMove={(event) => {
                              setHoveredNode({
                                x: event.clientX,
                                y: event.clientY,
                                name: node.data.name,
                                fullPath: node.data.fullPath,
                                timeText: node.data.timeText,
                                id: node.data.id,
                              });
                            }}
                            onMouseLeave={() => setHoveredNode(null)}
                            className="cursor-pointer"
                          >
                            <rect
                              x={node.x0}
                              y={node.y0}
                              width={nodeWidth}
                              height={nodeHeight}
                              fill={
                                isFolder
                                  ? "#f0f0f0"
                                  : colorScale(`${i % colors.length}`)
                              }
                              stroke={isHovered ? "#2563eb" : "#fff"}
                              strokeWidth={isHovered ? 3 : 2}
                              opacity={isHovered ? 1 : 0.9}
                              className="transition-all duration-150"
                            />

                            {showText && (
                              <text
                                x={node.x0 + nodeWidth / 2}
                                y={node.y0 + nodeHeight / 2}
                                textAnchor="middle"
                                fontSize={fontSize}
                                fontFamily="'JetBrains Mono', 'Consolas', 'Monaco', 'Courier New', monospace"
                                className="font-medium pointer-events-none"
                                fill={isFolder ? "#555" : "#000"}
                              >
                                <tspan x={node.x0 + nodeWidth / 2} dy="-0.6em">
                                  {truncate(node.data.name, {
                                    length: Math.floor(nodeWidth / 8),
                                  })}
                                </tspan>
                                {!isFolder && node.data.timeText && (
                                  <tspan
                                    x={node.x0 + nodeWidth / 2}
                                    dy="1.2em"
                                    fontSize={fontSize * 0.9}
                                  >
                                    {node.data.timeText}
                                  </tspan>
                                )}
                              </text>
                            )}
                          </Group>
                        );
                      })}
                  </Group>
                )}
              </Treemap>
            </svg>
          );
        }}
      </ParentSize>
    </div>
  );
};

export default FileActivityTreemapVisx;
