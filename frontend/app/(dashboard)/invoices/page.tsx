import { Invoice } from "@/lib/types";
import { fetchData, getSession } from "@/actions";
import { InvoicesTable } from "@/components/invoice-table";
import { ClientsApiResponse } from "@/components/clients-table";

export default async function Invoices() {
  const session = await getSession();

  const clients = await fetchData<ClientsApiResponse>(
    "compat/wakatime/v1/users/current/clients"
  );
  const invoices = await fetchData<{ data: Invoice[] }>(
    "compat/wakatime/v1/users/current/invoices"
  );

  return (
    <div className="panel panel-default p-2 px-6 my-6 mx-2 min-h-screen">
      <InvoicesTable
        clients={clients?.data || []}
        invoices={invoices?.data || []}
        token={session.token}
      />
    </div>
  );
}
