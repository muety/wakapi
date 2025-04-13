"use client";

import { format } from "date-fns";
import { LucidePlusCircle, LucideTrash2 } from "lucide-react";
import React from "react";

import { ApiClient } from "@/actions/api";
import { getCurrencySymbol } from "@/lib/constants/currencies";
import { Invoice, InvoiceLineItem } from "@/lib/types";
import { cn, formatNumber, getHours } from "@/lib/utils";

import { Icons } from "../icons";
import { Button } from "../ui/button";
import { toast } from "../ui/use-toast";
import styles from "./invoice-manager.module.css";
import { InvoicePreview } from "./invoice-preview";

interface iProps {
  data: Invoice;
}

export function InvoiceManager({ data }: iProps) {
  const date = new Date();
  const { client } = data;
  const [loading, setLoading] = React.useState(false);

  const [lineItems, setLineItems] = React.useState<InvoiceLineItem[]>(
    data.line_items.map((line_item) => ({
      ...line_item,
      total: getHours(line_item.total_seconds),
    }))
  );
  const defaultInvoiceSubtitle = () => {
    return `Invoice for the month of ${format(date, "MMMM yyyy")}.`;
  };
  const [origin, setFrom] = React.useState(data.origin);
  const [destination, setDestination] = React.useState(
    data.destination || client.name
  );
  const [tax, setTax] = React.useState(data.tax || "");
  const [heading, setPreamble] = React.useState(data.heading);
  const [finalMessage, setMainMessage] = React.useState(data.final_message);
  const [invoiceSummary, setInvoiceSummary] = React.useState(
    data.invoice_summary || defaultInvoiceSubtitle()
  );
  const [preview, setPreview] = React.useState(true);
  const [refreshIndex, setRefreshIndex] = React.useState(0);

  const totalInvoice = React.useMemo(() => {
    return lineItems.reduce((acc, item) => {
      return acc + getHours(+item.total_seconds) * client.hourly_rate;
    }, 0);
  }, [lineItems, client.hourly_rate]);

  const taxTotal = React.useMemo(() => {
    const parsedTax = parseInt(tax.toString());
    if (isNaN(parsedTax)) {
      return 0;
    }
    return totalInvoice * (parsedTax / 100);
  }, [totalInvoice, tax]);

  const netTotal = React.useMemo(() => {
    return totalInvoice + taxTotal;
  }, [totalInvoice, taxTotal]);

  const currencySymbol = getCurrencySymbol(client.currency);

  const deleteInvoiceItem = (index: number) => () => {
    setLineItems(lineItems.filter((_, i) => i !== index));
  };

  const addNewItem = () => {
    setLineItems([
      ...lineItems,
      {
        title: "",
        total_seconds: 0,
        auto_generated: false,
      } as any,
    ]);
  };

  const resourceUrl = `/v1/users/current/invoices/${data.id}`;

  const saveInvoice = async () => {
    try {
      const payload: Record<string, string | number | InvoiceLineItem[]> = {
        origin,
        destination,
        heading,
        final_message: finalMessage,
        invoice_summary: invoiceSummary,
        line_items: lineItems,
      };
      if (tax) {
        payload.tax = +tax;
      }
      setLoading(true);
      const response = await ApiClient.PUT(resourceUrl, payload);

      if (!response.success) {
        toast({
          title: "Failed to update invoice",
          variant: "destructive",
          description: "Please try again later",
        });
      } else {
        toast({
          title: "Invoice Saved",
          variant: "success",
        });
        setPreview(true);
      }
    } finally {
      setLoading(false);
    }
  };

  if (preview) {
    return (
      <InvoicePreview
        data={{
          ...data,
          line_items: lineItems,
          heading,
          final_message: finalMessage,
          invoice_summary: invoiceSummary,
          tax: +tax,
        }}
        onTogglePreview={() => setPreview(!preview)}
      />
    );
  }

  return (
    <div className={cn(styles.root, "px-6 my-6 mx-2 min-h-screen")}>
      <h1 className="text my-5 text-xl">
        Invoice for <b>{client.name}</b>
      </h1>
      <main className={styles.main}>
        <div className="flex justify-between">
          <div className="w-100 w-full max-w-lg">
            <div>
              <h1 className="text-3xl">INVOICE</h1>
              <textarea
                rows={2}
                className={styles.invoiceInput}
                placeholder="Invoice Subtitle"
                defaultValue={invoiceSummary}
                onChange={(event) => setInvoiceSummary(event.target.value)}
              />
            </div>

            <div className="my-4">
              <label htmlFor="from">From</label>
              <textarea
                rows={4}
                className={styles.invoiceInput}
                placeholder="Invoice Item"
                defaultValue={origin}
                onChange={(event) => setFrom(event.target.value)}
              />
            </div>

            <div className="my-4">
              <label htmlFor="to">To</label>
              <textarea
                id="to"
                rows={4}
                className={styles.invoiceInput}
                placeholder="Billing address"
                defaultValue={destination}
                onChange={(event) => setDestination(event.target.value)}
              />
            </div>
          </div>
          <div className="flex">
            <div className="mr-1 flex flex-col items-end justify-items-end">
              <h1 className="font-bold">Invoice #: </h1>
              <h1 className="font-bold">Date: </h1>
            </div>
            <div>
              <p>INV-0001</p>
              <p>{format(date, "MMM dd, yyyy")}</p>
            </div>
          </div>
        </div>
        <div className="mt-4">
          <label htmlFor="from">Preamble</label>
          <textarea
            className={cn(styles.invoiceInput)}
            placeholder="Write something here"
            defaultValue={heading}
            onChange={(event) => setPreamble(event.target.value)}
          ></textarea>
        </div>
        {/* <div className={cn(styles.newTable, "p-2")}> */}
        <div className={cn(styles.invoiceTableWrapper, "shadow")}>
          <table className={cn(styles.invoiceTable, "")}>
            <thead>
              <tr className={cn(styles.invoiceRow, styles.invoiceHeader)}>
                <th>Item</th>
                <th>Price ({currencySymbol})</th>
                <th>Qty (Hrs)</th>
                <th>Amount ({currencySymbol})</th>
                <th className={styles.invoiceRowAction}></th>
              </tr>
            </thead>
            <tbody style={{ borderRadius: "0.5rem" }}>
              {lineItems.map((item, index) => (
                <tr key={index} className={cn(styles.invoiceRow)}>
                  <td>
                    <input
                      className={cn(styles.invoiceInput, styles.inputOutlined)}
                      placeholder="Invoice Item"
                      defaultValue={item.title}
                      onChange={(event) => (item.title = event.target.value)}
                    />
                  </td>
                  <td>{client.hourly_rate.toFixed(2)}</td>
                  <td>
                    {!(item as InvoiceLineItem).auto_generated ? (
                      <input
                        className={cn(styles.invoiceInput, "text-center")}
                        placeholder="Invoice Item"
                        defaultValue={getHours(item.total_seconds || 0)}
                        onChange={(event) => {
                          item.total_seconds =
                            parseFloat(event.target.value) * 3600;
                          setRefreshIndex(refreshIndex + 1);
                        }}
                      />
                    ) : (
                      getHours(item.total_seconds).toFixed(2)
                    )}
                  </td>
                  <td>
                    {/* How do I make the computation below reactive to the change in the total? */}
                    {/* {(item.total_seconds)} */}
                    {isNaN(parseFloat(String(item.total_seconds)))
                      ? 0
                      : (
                          getHours(item.total_seconds) * client.hourly_rate
                        ).toFixed(2)}
                  </td>
                  <td className={cn(styles.invoiceAction, "text-right")}>
                    <div className="flex justify-end">
                      <div
                        className="m-0 flex justify-center rounded-sm border border-red-400 p-0 text-red-400 hover:border-red-700 hover:text-red-700"
                        style={{ width: "25px", padding: "2px" }}
                      >
                        <LucideTrash2
                          onClick={deleteInvoiceItem(index)}
                          className="m-0 size-4 p-0 text-center "
                        />
                      </div>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr className={cn(styles.invoiceRow)}>
                <td onClick={addNewItem}>
                  <span
                    className={cn(styles.addItem, "border border-slate-300")}
                  >
                    <LucidePlusCircle className="size-4" />
                    <span className="ml-2">Add item</span>
                  </span>
                </td>
                <td colSpan={3}></td>
              </tr>
            </tfoot>
          </table>
        </div>

        <div className="my-3 flex justify-end">
          <div className="flex gap-3">
            <div className="mr-1 flex flex-col items-end justify-center gap-1">
              <h1 className="font-bold">Total </h1>
              <h1 className="font-bold">
                Tax{" "}
                <input
                  className={cn(styles.invoiceInput, styles.taxInput)}
                  defaultValue={tax}
                  onInput={(e) => setTax(e.currentTarget.value)}
                />
                (%)
              </h1>
              <h1 className="font-bold">Total After Tax </h1>
            </div>
            <div className="flex flex-col items-end justify-end gap-1">
              <p>{formatNumber(totalInvoice, { currency: client.currency })}</p>
              <p>{formatNumber(taxTotal, { currency: client.currency })}</p>
              <p>{formatNumber(netTotal, { currency: client.currency })}</p>
            </div>
          </div>
        </div>
        {/* </div> */}

        <textarea
          className={cn(styles.invoiceInput, "my-5")}
          placeholder="Message at the bottom of the invoice"
          defaultValue={finalMessage}
          onChange={(event) => setMainMessage(event.target.value)}
        ></textarea>
        <div className={styles.invoiceFooter}>
          <Button onClick={saveInvoice}>
            {loading ? (
              <div className="flex items-center gap-2">
                <Icons.spinner className="animate-spin" />
                Saving...
              </div>
            ) : (
              "Save Invoice"
            )}
            {/* Save Invoice */}
          </Button>
        </div>
      </main>
    </div>
  );
}
