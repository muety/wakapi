import { Metadata } from "next";

import { Invoice } from "@/lib/types";
import { fetchData } from "@/actions";
import { InvoicesTable } from "@/components/invoice-table";
import { ClientsApiResponse } from "@/components/clients-table";

export const metadata: Metadata = {
  title: "Invoices",
  description: "Wakana invoices, create and track invoices for billable hours.",
};

export default async function Invoices() {
  const clients = await fetchData<ClientsApiResponse>(
    "/v1/users/current/clients"
  );
  const invoices = await fetchData<{ data: Invoice[] }>(
    "/v1/users/current/invoices"
  );

  return (
    <div className="my-6">
      <div className="mb-5 flex items-center justify-start">
        <h1 className="text-4xl">Invoices</h1>
      </div>
      <InvoicesTable
        clients={clients?.data || []}
        invoices={invoices?.data || []}
      />
    </div>
  );
}
