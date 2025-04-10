"use client";

import { LucidePlus } from "lucide-react";
import { useRouter } from "next/navigation";
import React from "react";

import { Client } from "./clients-table";
import { NewInvoiceForm } from "./forms/new-invoice-form";
import { Button } from "./ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./ui/dialog";
import { toast } from "./ui/use-toast";
import { postData } from "@/actions/api";
import { get } from "lodash";
import { Invoice } from "@/lib/types";

export interface iProps {
  clients: Client[];
  open?: boolean;
  onChange: (open: boolean) => void;
}

export function AddInvoice({ clients, onChange, open }: iProps) {
  const [loading, setLoading] = React.useState(false);

  const router = useRouter();

  const createInvoice = async (values: Record<string, any>) => {
    try {
      const { client, start_date, end_date } = values;

      const resourceUrl = `/v1/users/current/invoices`;
      setLoading(true);

      const response = await postData<{ data: Invoice }>(resourceUrl, {
        client_id: client,
        start_date: start_date.toISOString(),
        end_date: end_date.toISOString(),
      });

      if (!response.success) {
        console.log("[response]", response);
        toast({
          title: "Failed to create client",
          variant: "destructive",
        });
      } else {
        toast({
          title: "Invoice created successfully",
          variant: "success",
        });
        router.push(`/invoices/${response.data.data.id}`);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(open) => onChange(open)}>
      <DialogTrigger asChild>
        <Button
          onClick={() => onChange(true)}
          variant="outline"
          className="w-48 bg-black text-white"
        >
          <LucidePlus className="size-4" />
          Create Invoice
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[625px]">
        <DialogHeader>
          <DialogTitle className="dialog-label">Create Invoice</DialogTitle>
          <DialogDescription className="sr-only">
            Create Client
          </DialogDescription>
        </DialogHeader>
        <NewInvoiceForm
          clients={clients}
          onSubmit={createInvoice}
          loading={loading}
        />
      </DialogContent>
    </Dialog>
  );
}
