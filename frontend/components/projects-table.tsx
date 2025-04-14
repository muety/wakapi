"use client";
import Fuse from "fuse.js";
import * as React from "react";

import { Input } from "@/components/ui/input";

import ProjectListCard from "./project-list-card";

export type Project = {
  id: string;
  name: string;
  last_heartbeat_at: string;
  created_at: string;
  urlencoded_name: string;
  human_readable_last_heartbeat_at: string;
};

export type ProjectsApiResponse = {
  data: Project[];
};

export function ProjectsTable({ projects }: { projects: Project[] }) {
  const [searchQuery, setSearchQuery] = React.useState("");

  const fuse = React.useMemo(
    () =>
      new Fuse(projects, {
        keys: ["name", "id", "urlencoded_name"], // Fields to search
        threshold: 0.3,
        findAllMatches: true,
      }),
    [projects]
  );

  const filteredProjects = React.useMemo(() => {
    if (!searchQuery) return projects; // Return all items if no search query
    return fuse.search(searchQuery).map((result) => result.item);
  }, [searchQuery, fuse, projects]);

  return (
    <div className="w-full">
      <div className="flex items-center py-4">
        <Input
          placeholder="Filter projects"
          value={searchQuery}
          onChange={(event) => setSearchQuery(event.target.value)}
        />
      </div>
      <div className="flex flex-wrap gap-1 space-y-3">
        {filteredProjects.map((p) => (
          <ProjectListCard key={p.id} project={p} />
        ))}
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm">
          showing {filteredProjects.length} results.
        </div>
      </div>
    </div>
  );
}
