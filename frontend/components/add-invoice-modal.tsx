"use client";

import React from "react";

import { Dialog } from "@/components/ui/dialog";

export default function AddInvoiceDialog() {
  const [dialogOpen, setDialogOpen] = React.useState(false);
  return (
    <Dialog open={dialogOpen} onOpenChange={(open) => setDialogOpen(open)}>
      {/* Start by showing them a way to select client */}
      {/* Then proceed to the invoice view where you show them the content of the invoice */}
      {/* Finally let them download? */}
    </Dialog>
  );
}
