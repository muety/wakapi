"use client";

import { format } from "date-fns";
import { LucideEdit } from "lucide-react";
import React from "react";

import { getCurrencySymbol } from "@/lib/constants/currencies";
import { Invoice } from "@/lib/types";
import { cn, formatCurrency, formatNumber, getHours } from "@/lib/utils";

import { InvoicePDF } from "../pdf/pdf-export";
import { RawHTML } from "../raw-html";
import { Button } from "../ui/button";
import styles from "./invoice-manager.module.css";

interface iProps {
  data: Invoice;
  onTogglePreview: () => void;
}

export function InvoicePreview({ data, onTogglePreview }: iProps) {
  const { client, tax, line_items } = data;

  const totalInvoice = React.useMemo(() => {
    return line_items.reduce((acc, item) => {
      return acc + getHours(item.total_seconds) * client.hourly_rate;
    }, 0);
  }, [line_items]);

  const taxTotal = React.useMemo(() => {
    if (isNaN(tax)) {
      return 0;
    }
    return totalInvoice * (tax / 100);
  }, [totalInvoice, tax]);

  const netTotal = React.useMemo(() => {
    return totalInvoice + taxTotal;
  }, [totalInvoice, taxTotal]);

  const currencySymbol = getCurrencySymbol(client.currency) || client.currency;

  return (
    <div className={cn(styles.root, "px-6 my-6 mx-2 min-h-screen")}>
      <main id="target-invoice" className={cn(styles.mainPreview, "")}>
        <div className="mb-3 flex justify-end gap-1">
          <Button
            size={"sm"}
            className="size-7 bg-white p-1 hover:bg-white hover:opacity-70"
            variant="outline"
            onClick={onTogglePreview}
          >
            <LucideEdit className="size-4 text-black" />
          </Button>
          <InvoicePDF invoiceData={data} />
        </div>
        <div className="flex justify-between">
          <div className="w-100 max-w-lg">
            <div>
              <h1 className="text-3xl">INVOICE</h1>
              <RawHTML source={data.invoice_summary} fallback="" />
            </div>
            <div className="my-2 mt-4">
              <label htmlFor="from">From</label>
              <RawHTML source={data.origin} fallback="" />
            </div>{" "}
            <br />
            <div className="my-2">
              <label htmlFor="to">Bill To</label>
              <RawHTML source={data.destination} fallback="" />
            </div>
          </div>
          <div>
            <div className="flex">
              <div className="mr-1 flex flex-col items-end justify-items-end">
                <h1 className="font-bold">Invoice #: </h1>
                <h1 className="font-bold">Date: </h1>
              </div>
              <div>
                <p>{data.id.slice(0, 9)}</p>
                <p>{format(data.created_at, "MMM dd, yyyy")}</p>
              </div>
            </div>
          </div>
        </div>
        {data.heading && (
          <div className="my-4 mt-8">
            <RawHTML source={data.heading} fallback="" />
          </div>
        )}
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
              {line_items.map((item, index) => (
                <tr key={index} className={cn(styles.invoiceRow)}>
                  <td>{item.title}</td>
                  <td>{client.hourly_rate.toFixed(2)}</td>
                  <td>{getHours(item.total_seconds).toFixed(2)}</td>
                  <td>
                    {formatNumber(
                      getHours(item.total_seconds) * client.hourly_rate,
                      {
                        currency: client.currency,
                        style: "currency",
                      }
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="mt-12 flex justify-end">
          <div className="flex gap-4">
            <div className="mr-1 flex flex-col items-end justify-center gap-1 text-lg">
              <h1 className="font-semibold">Total </h1>
              <h1 className="font-semibold">Tax ({tax || ""}%) </h1>
              <h1 className="font-semibold">Net Total </h1>
            </div>
            <div className="flex flex-col items-end justify-end gap-1 text-lg">
              <p>{formatCurrency(totalInvoice, currencySymbol)}</p>
              <p>{formatCurrency(taxTotal, currencySymbol)}</p>
              <p>{formatCurrency(netTotal, currencySymbol)}</p>
            </div>
          </div>
        </div>

        <RawHTML source={data.final_message} fallback="" />
      </main>
    </div>
  );
}
