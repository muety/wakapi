"use client";

import { formatDistanceToNow } from "date-fns";
import { startCase, truncate } from "lodash";
import Link from "next/link";

import { Card, CardContent, CardHeader } from "@/components/ui/card";

import { ProjectTimeImage } from "./project-time-image";

interface ProjectData {
  id: string;
  name: string;
  last_heartbeat_at: string;
  human_readable_last_heartbeat_at: string;
  urlencoded_name: string;
  created_at: string;
}

export default function ProjectListCard({ project }: { project: ProjectData }) {
  // Calculate last updated time
  const lastUpdatedAt = formatDistanceToNow(
    new Date(project.last_heartbeat_at),
    { addSuffix: true }
  );

  // Calculate project age
  const projectAge = formatDistanceToNow(new Date(project.created_at), {
    addSuffix: false,
  });

  // Generate a random color based on project name (for consistency)
  const getRandomColor = (name: string) => {
    // Generate a hash from the string
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }

    // Convert to hex color
    let color = "#";
    for (let i = 0; i < 3; i++) {
      const value = (hash >> (i * 8)) & 0xff;
      color += ("00" + value.toString(16)).substr(-2);
    }

    return color;
  };

  // Get initials from project name
  const getProjectInitials = (name: string) => {
    return name
      .split("-")
      .map((word) => word[0]?.toUpperCase() || "")
      .join("")
      .substring(0, 2);
  };

  const displayName = startCase(truncate(project.name, { length: 20 }));
  const projectId = truncate(project.name, { length: 15 });

  return (
    <Link
      href={`/projects/${project.id}`}
      className="md:w-full card-border border-border"
    >
      <Card className="p-1 w-full overflow-hidden border border-border/40 shadow-sm duration-300 ease-in-out hover:border-white/15 hover:bg-white/[4%] cursor-pointer">
        <CardHeader className="relative pb-2">
          <div className="flex justify-between items-start">
            <div className="flex items-center space-x-3">
              <div
                className="h-10 w-10 rounded-sm flex items-center justify-center text-white font-semibold flex-shrink-0"
                style={{ backgroundColor: getRandomColor(project.name) }}
              >
                {getProjectInitials(project.name)}
              </div>
              <div>
                <h3
                  className="text-xl font-semibold tracking-tight"
                  title={startCase(project.name)}
                >
                  {displayName}
                </h3>
                <p className="text-sm text-muted-foreground">
                  Created {projectAge} ago
                </p>
              </div>
            </div>
            <ProjectTimeImage project={project} />
          </div>
        </CardHeader>
        <CardContent className="pb-2">
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Project ID</span>
              <span className="font-mono">{projectId}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Last updated at</span>
              <span>{lastUpdatedAt}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
