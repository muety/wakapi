import { PDFDownloadLink } from "@react-pdf/renderer";
import { LucidePrinter } from "lucide-react";

import { Invoice } from "@/lib/types";

import { Button } from "../ui/button";
import { InvoicePDFViewer } from "./invoice-pdf-viewer";

interface iProps {
  invoiceData: Invoice;
}

export const InvoicePDF = (props: iProps) => {
  return (
    <Button
      size={"sm"}
      variant="outline"
      className="size-7 bg-white p-1 hover:bg-white hover:opacity-70"
    >
      <PDFDownloadLink
        fileName="invoice.pdf"
        document={<InvoicePDFViewer invoiceData={props.invoiceData} />}
      >
        <LucidePrinter className="size-4 text-black" />
      </PDFDownloadLink>
    </Button>
  );
};
