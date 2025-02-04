import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export function InvoiceSkeleton() {
  return (
    <Table>
      <TableCaption>A list of your recent invoices</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Client</TableHead>
          <TableHead>Billable Hours</TableHead>
          <TableHead>Amount</TableHead>
          <TableHead className="text-right">Currency</TableHead>
          <TableHead className="text-right">Duration</TableHead>
          <TableHead className="text-right">Created</TableHead>
          <TableHead className="text-right">Updated</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: 5 }).map((_, index) => (
          <TableRow key={index}>
            <TableCell className="font-medium">
              <Skeleton className="h-4 w-[80px]" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-[60px]" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-[80px]" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-[80px]" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-[80px]" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-[80px]" />
            </TableCell>
            <TableCell className="text-right">
              <Skeleton className="h-4 w-[60px] ml-auto" />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
