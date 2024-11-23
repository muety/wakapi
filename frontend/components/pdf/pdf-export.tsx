import React from "react";

import { Button } from "../ui/button";
import { LucidePrinter } from "lucide-react";
import { PDFDownloadLink } from "@react-pdf/renderer";
import { InvoicePDFViewer } from "./invoice-pdf-viewer";
import { Invoice } from "@/lib/types";

interface iProps {
  invoiceData: Invoice;
}

export const InvoicePDF = (props: iProps) => {
  return (
    <Button
      size={"sm"}
      variant="outline"
      className="bg-white hover:bg-white hover:opacity-70 p-1 h-7 w-7"
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
