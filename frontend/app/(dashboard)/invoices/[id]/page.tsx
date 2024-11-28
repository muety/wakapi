import { fetchData } from "@/actions";
import { InvoiceManager } from "@/components/invoice/invoice-manager";
import { Invoice } from "@/lib/types";

interface iProps {
  searchParams: Record<string, any>;
  params: { id: string };
}

export default async function InvoiceDetail({ params }: iProps) {
  const invoiceData = await fetchData<{ data: Invoice }>(
    `compat/wakatime/v1/users/current/invoices/${params.id}`
  );

  if (!invoiceData) {
    return (
      <div>
        Unexpected error fetching invoice data. Api server might be down.
      </div>
    );
  }

  return <InvoiceManager data={invoiceData?.data} />;
}
