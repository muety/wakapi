import { fetchData } from "@/actions";
import { InvoiceManager } from "@/components/invoice/invoice-manager";
import { Invoice } from "@/lib/types";
import { notFound } from "next/navigation";

interface iProps {
  searchParams: Record<string, any>;
  params: { id: string };
}

export default async function InvoiceDetail({ params }: iProps) {
  const invoiceData = await fetchData<{ data: Invoice }>(
    `compat/wakatime/v1/users/current/invoices/${params.id}`
  );

  if (!invoiceData || !invoiceData.data) {
    notFound();
  }

  return <InvoiceManager data={invoiceData?.data} />;
}
