import { fetchData, getSession } from "@/actions";
import { ClientsApiResponse } from "@/components/clients-table";
import { InvoicesTable } from "@/components/invoice-table";
import { Invoice } from "@/lib/types";

export default async function Invoices() {
  const session = await getSession();

  const clients = await fetchData<ClientsApiResponse>(
    "compat/wakatime/v1/users/current/clients"
  );
  const invoices = await fetchData<{ data: Invoice[] }>(
    "compat/wakatime/v1/users/current/invoices"
  );

  return (
    <div className="panel panel-default mx-2 my-6 min-h-screen p-2 px-6">
      <InvoicesTable
        clients={clients?.data || []}
        invoices={invoices?.data || []}
        token={session.token}
      />
    </div>
  );
}
