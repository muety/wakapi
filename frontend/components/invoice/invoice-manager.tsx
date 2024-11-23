"use client";

import React from "react";
import styles from "./invoice-manager.module.css";

import { format } from "date-fns";
import { Button } from "../ui/button";
import { InvoicePreview } from "./invoice-preview";
import { Invoice, InvoiceLineItem } from "@/lib/types";
import { cn, formatNumber, getHours } from "@/lib/utils";
import { LucidePlusCircle, LucideTrash2 } from "lucide-react";
import { getCurrencySymbol } from "@/lib/constants/currencies";
import { NEXT_PUBLIC_API_URL } from "@/lib/constants/config";
import { toast } from "../ui/use-toast";
import { Icons } from "../icons";
import { useClientSession } from "@/lib/session";

interface iProps {
  data: Invoice;
}

export function InvoiceManager({ data }: iProps) {
  const session = useClientSession();
  const token = React.useMemo(
    () => session.data?.token || "",
    [session.data?.token]
  );
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
  }, [lineItems]);

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

  const resourceUrl = `${NEXT_PUBLIC_API_URL}/api/compat/wakatime/v1/users/current/invoices/${data.id}`;

  const saveInvoice = async () => {
    try {
      const payload: Record<string, any> = {
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
      const response = await fetch(resourceUrl, {
        method: "PUT",
        body: JSON.stringify(payload),
        headers: {
          accept: "application/json",
          "content-type": "application/json",
          token: `${token}`,
        },
      });

      if (!response.ok) {
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
      <h1 className="my-5 text-xl text">
        Invoice for <b>{client.name}</b>
      </h1>
      <main className={styles.main}>
        <div className="flex justify-between">
          <div className="max-w-lg w-100 w-full">
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
            <div className="flex flex-col justify-items-end items-end mr-1">
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
                    {!(item as any).auto_generated ? (
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
                        className="flex hover:border-red-700 hover:text-red-700 text-red-400 justify-center border-red-400 border rounded-sm p-0 m-0"
                        style={{ width: "25px", padding: "2px" }}
                      >
                        <LucideTrash2
                          onClick={deleteInvoiceItem(index)}
                          className="size-4 text-center m-0 p-0 "
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

        <div className="flex justify-end my-3">
          <div className="flex gap-3">
            <div className="flex flex-col justify-center items-end mr-1 gap-1">
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
            <div className="flex flex-col justify-end items-end gap-1">
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
