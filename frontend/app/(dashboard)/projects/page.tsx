import {
  ProjectsApiResponse,
  ProjectsTable,
} from "@/components/projects-table";
import { fetchData } from "@/actions";

export default async function Projects({ searchParams }: Record<string, any>) {
  const projects = await fetchData<ProjectsApiResponse | null>(
    `compat/wakatime/v1/users/current/projects${new URLSearchParams(
      searchParams
    )}`
  );

  return (
    <div className="panel panel-default p-2 px-6 my-6 mx-2">
      <ProjectsTable projects={projects?.data || []} />
    </div>
  );
}
