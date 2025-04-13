import { Metadata } from "next";

import { fetchData } from "@/actions";
import { Project } from "@/components/projects-table";
import { ClientsApiResponse, ClientsTable } from "@/components/clients-table";

export const metadata: Metadata = {
  title: "Clients",
  description: "Wakana clients, manage your freelance clients.",
};

export default async function Clients() {
  const clients = await fetchData<ClientsApiResponse | null>(
    "/v1/users/current/clients"
  );
  const projects = await fetchData<{ data: Project[] } | null>(
    "/v1/users/current/projects"
  );

  return (
    <div className="my-6">
      <div className="mb-5 flex items-center justify-start">
        <h1 className="text-4xl">Clients</h1>
      </div>
      <ClientsTable
        clients={clients?.data || []}
        projects={projects?.data || []}
      />
    </div>
  );
}
