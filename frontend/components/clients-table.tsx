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
import { humanizeDate } from "@/lib/utils";

import { AddClient } from "./add-client";
import { Project } from "./projects-table";
import { Confirm } from "./ui/confirm";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import { toast } from "./ui/use-toast";

export type Client = {
  id: string;
  name: string;
  user_id?: string;
  projects: string[];
  hourly_rate: number;
  currency: string;
  updated_at: string;
  created_at: string;
};

export type ClientsApiResponse = {
  data: Client[];
};

export function ClientsTable({
  clients: currentClients,
  projects,
  token,
}: {
  clients: Client[];
  projects: Project[];
  token: string;
}) {
  const [editing, setEditing] = React.useState<Client | null>(null);
  const [deleting, setDeleting] = React.useState<Client | null>(null);
  const [clients, setClients] = React.useState<Client[]>(currentClients);
  const [showClientModal, setShowClientModal] = React.useState<boolean>(false);
  const [loading, setLoading] = React.useState<boolean>(false);

  const deleteClient = async () => {
    try {
      if (!deleting) {
        return;
      }
      const resourceUrl = `${NEXT_PUBLIC_API_URL}/api/v1/users/current/clients/${deleting.id}`;

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
          description: `Client: ${deleting.name} - deleted`,
          variant: "success",
        });
        setClients(clients.filter((client) => client.id !== deleting?.id));
        setDeleting(null);
      }
    } finally {
      setLoading(false);
    }
  };

  const editRow = (row: Client) => {
    setEditing(row);
    setShowClientModal(true);
  };

  const deleteRow = (row: Client) => {
    setDeleting(row);
  };

  const columns: ColumnDef<Client>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <a href="#">{row.getValue("name")}</a>,
    },
    {
      accessorKey: "projects",
      header: "Projects",
    },
    {
      accessorKey: "currency",
      header: "Currency",
    },
    {
      accessorKey: "hourly_rate",
      header: "Rate/Hr",
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
              <DropdownMenuLabel>Actions</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => editRow(row.original)}>
                Edit
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
    data: clients,
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
          <AddClient
            onAdd={(newClient: Client) => {
              setClients([...clients, newClient]);
              setEditing(null);
            }}
            onEdit={(editedClient: Client) => {
              const index = clients.findIndex(
                (client) => client.id === editedClient.id
              );
              if (index !== -1) {
                const updatedClients = [...clients];
                updatedClients.splice(index, 1, editedClient);
                setClients(updatedClients);
                setEditing(null);
              }
            }}
            onChange={(open) => {
              setShowClientModal(open);
              if (!open) {
                setEditing(null);
              }
            }}
            token={token}
            projects={projects as any}
            editing={editing}
            open={showClientModal}
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
                  className="h-24 text-center"
                >
                  You have no clients.
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
        title="Delete Client"
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
