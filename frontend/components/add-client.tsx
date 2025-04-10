"use client";

import React from "react";

import { Project } from "@/lib/types";
import { LucidePlus } from "lucide-react";

import { Button } from "./ui/button";
import { toast } from "./ui/use-toast";
import { Client } from "./clients-table";
import { ClientForm } from "./forms/client-form";
import { postData, updateData } from "@/actions/api";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "./ui/dialog";

export interface iProps {
  onAdd: (client: any) => void;
  onEdit: (client: any) => void;
  projects: Project[];
  editing: Client | null;
  open?: boolean;
  onChange: (open: boolean) => void;
}

export function AddClient({
  projects,
  onAdd,
  onEdit,
  editing,
  onChange,
  open,
}: iProps) {
  const submitHandler = (values: any) => {
    // type me
    console.log("values", values);
    if (editing) {
      updateClient(values, editing.id);
    } else {
      createClient(values);
    }
  };
  const [loading, setLoading] = React.useState(false);

  const updateClient = async (values: Partial<Client>, id: string) => {
    try {
      const resourceUrl = `/v1/users/current/clients/${id}`;
      setLoading(true);
      const response = await updateData(resourceUrl, values);

      if (!response.success) {
        const data = response.data;
        toast({
          title: response.error?.message || "Failed to update client",
          variant: "destructive",
        });
      } else {
        if (onEdit) {
          const data = response.data;
          onEdit(data.data);
        }
        toast({
          title: "Title Updated",
          variant: "success",
        });
        onChange(false);
      }
    } finally {
      setLoading(false);
    }
  };

  const createClient = async (values: Partial<Client>) => {
    try {
      const resourceUrl = `/v1/users/current/clients`;
      setLoading(true);
      const response = await postData(resourceUrl, values);

      if (!response.success) {
        toast({
          title: "Failed to create client",
          variant: "destructive",
        });
      } else {
        const data = response.data;
        onAdd(data.data);
        toast({
          title: "Client Created",
          variant: "success",
        });
        onChange(false);
      }
      onChange(false);
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
          className="w-36 bg-black text-white"
        >
          <LucidePlus className="size-4" />
          New Client
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[625px]">
        <DialogHeader className="">
          <DialogTitle className="text-2xl">Client</DialogTitle>
          <DialogDescription className="sr-only">
            Create Client
          </DialogDescription>
        </DialogHeader>
        <ClientForm
          projects={projects as any}
          onSubmit={submitHandler}
          editing={editing}
          loading={loading}
        />
      </DialogContent>
    </Dialog>
  );
}
