"use client";

import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import {
  ColumnDef,
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
  VisibilityState,
} from "@tanstack/react-table";
import { format } from "date-fns";
import { LucidePlus } from "lucide-react";
import Link from "next/link";
import * as React from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { NEXT_PUBLIC_API_URL } from "@/lib/constants/config";
import { Invoice, InvoiceLineItem } from "@/lib/types";
import { convertSecondsToHours, formatNumber, humanizeDate } from "@/lib/utils";

import { AddInvoice } from "./add-invoice";
import { Client } from "./clients-table";
import { Confirm } from "./ui/confirm";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import { toast } from "./ui/use-toast";

export function InvoicesTable({
  clients,
  invoices: currentInvoices = [],
  token,
}: {
  clients: Client[];
  invoices: Invoice[];
  token: string;
}) {
  const [deleting, setDeleting] = React.useState<Invoice | null>(null);
  const [invoices, setInvoices] = React.useState<Invoice[]>([]);
  const [showInvoiceModal, setShowInvoiceModal] =
    React.useState<boolean>(false);
  const [loading, setLoading] = React.useState<boolean>(false);

  React.useEffect(() => {
    setInvoices(currentInvoices);
  }, [currentInvoices]);

  const deleteClient = async () => {
    try {
      if (!deleting) {
        return;
      }
      const resourceUrl = `${NEXT_PUBLIC_API_URL}/api/compat/wakatime/v1/users/current/invoices/${deleting.id}`;

      setLoading(true);
      const response = await fetch(resourceUrl, {
        method: "DELETE",
        headers: {
          accept: "application/json",
          "content-type": "application/json",
          token: `${token}`,
        },
      });

      if (!response.ok) {
        toast({
          title: "Failed to delete goal",
          variant: "destructive",
        });
      } else {
        toast({
          title: "Deleted",
          description: `Invoice: ${deleting.name} - deleted`,
          variant: "success",
        });
        setInvoices(invoices.filter((client) => client.id !== deleting?.id));
        setDeleting(null);
      }
    } finally {
      setLoading(false);
    }
  };

  const deleteRow = (row: Invoice) => {
    setDeleting(row);
  };

  const columns: ColumnDef<Invoice>[] = [
    {
      accessorKey: "client.name",
      header: "Client",
    },
    {
      accessorKey: "hours",
      header: "Billable Hours",
      cell: ({ row }) => {
        const line_items = row.original.line_items as InvoiceLineItem[];
        const total_seconds = (line_items || []).reduce(
          (total: number, item: InvoiceLineItem) => total + item.total_seconds,
          0
        );

        return <div>{convertSecondsToHours(total_seconds)}</div>;
      },
    },
    {
      accessorKey: "amount",
      header: "Amount",
      cell: ({ row }) => {
        const line_items = row.original.line_items as InvoiceLineItem[];
        console.log("line_items", line_items, row);
        const total_seconds = (line_items || []).reduce(
          (total: number, item: InvoiceLineItem) => total + item.total_seconds,
          0
        );

        const total_hours = total_seconds / 60 / 60;
        const total_amount = total_hours * row.original.client.hourly_rate;

        return (
          <div>
            {formatNumber(total_amount, {
              currency: row.original.client.currency,
            })}
          </div>
        );
      },
    },
    {
      accessorKey: "client.currency",
      header: "Currency",
    },
    {
      accessorKey: "created_at",
      header: "Duration",
      cell: ({ row }) => {
        const dateFormat = "dd/MM/yyyy";
        const start_date = row.original.start_date;
        const end_date = row.original.end_date;

        const formattedStartDate = format(start_date, dateFormat);
        const formattedEndDate = format(end_date, dateFormat);
        return <div>{`${formattedStartDate} - ${formattedEndDate}`}</div>;
      },
    },
    {
      accessorKey: "created_at",
      header: "Created",
      cell: ({ row }) => <div>{humanizeDate(row.getValue("created_at"))}</div>,
    },
    {
      accessorKey: "updated_at",
      header: "Updated",
      cell: ({ row }) => <div>{humanizeDate(row.getValue("updated_at"))}</div>,
    },
    {
      id: "actions",
      enableHiding: false,
      cell: ({ row }) => {
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="size-8 p-0">
                <span className="sr-only">Open menu</span>
                <DotsHorizontalIcon className="size-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem>
                <Link href={`/invoices/${row.original.id}`}>View</Link>
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => deleteRow(row.original)}>
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});

  const table = useReactTable({
    data: invoices,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  });

  return (
    <div className="w-full">
      <div className="flex items-center justify-between gap-5 py-4">
        <Input
          placeholder="Filter clients"
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(event) =>
            table.getColumn("name")?.setFilterValue(event.target.value)
          }
          // className="bg-white text-black"
        />
        <div className="grow">
          <AddInvoice
            token={token}
            clients={clients}
            open={showInvoiceModal}
            onChange={(val) => setShowInvoiceModal(val)}
          />
        </div>
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id} className="font-bold">
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 pt-4 text-center"
                >
                  {/* No invoices. <br /> */}
                  <Button
                    onClick={() => setShowInvoiceModal(true)}
                    variant="outline"
                    className="my-3 w-48 bg-black text-white"
                  >
                    <LucidePlus className="size-4" />
                    Create Invoice
                  </Button>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm">
          showing {table.getFilteredRowModel().rows.length} results.
        </div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
      <Confirm
        title="Delete Invoice"
        description={`Delete client: ${deleting?.name}?`}
        onConfirm={() => deleteClient()}
        open={deleting !== null}
        onCancel={() => setDeleting(null)}
        loading={loading}
      >
        <></>
      </Confirm>
    </div>
  );
}
