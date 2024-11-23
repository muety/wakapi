import { Project } from "@/components/projects-table";
import { fetchData, getSession } from "@/actions";
import { ClientsApiResponse, ClientsTable } from "@/components/clients-table";

export default async function Clients() {
  const session = await getSession();

  const clients = await fetchData<ClientsApiResponse | null>(
    "compat/wakatime/v1/users/current/clients"
  );
  const projects = await fetchData<{ data: Project[] } | null>(
    "compat/wakatime/v1/users/current/projects"
  );

  return (
    <div className="panel panel-default p-2 px-6 my-6 mx-2 min-h-screen">
      <ClientsTable
        clients={clients?.data || []}
        projects={projects?.data || []}
        token={session.token}
      />
    </div>
  );
}
