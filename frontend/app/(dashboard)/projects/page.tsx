import { fetchData } from "@/actions";
import {
  ProjectsApiResponse,
  ProjectsTable,
} from "@/components/projects-table";

export default async function Projects({
  searchParams,
}: Record<string, string>) {
  const projects = await fetchData<ProjectsApiResponse | null>(
    `compat/wakatime/v1/users/current/projects${new URLSearchParams(
      searchParams
    )}`
  );

  return (
    <div className="panel panel-default mx-2 my-6 p-2 px-6">
      <ProjectsTable projects={projects?.data || []} />
    </div>
  );
}
