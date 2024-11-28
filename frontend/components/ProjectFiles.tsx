"use client";

import { truncate } from "lodash";
import { LucideCopy } from "lucide-react";

import { toast } from "@/components/ui/use-toast";
import { SummariesResponse } from "@/lib/types";

export interface iProps {
  data: SummariesResponse[];
  field: "branches" | "entities";
  title: string;
  showCopy?: boolean;
}

const formatFileName = (fileName: string) => {
  const arr = Array.from(fileName);
  const reversed = truncate(arr.reverse().join(""), { length: 30 });
  return Array.from(reversed).reverse().join("");
};

function TableRow({
  name,
  text,
  showCopy,
}: {
  name: string;
  text: string;
  showCopy: boolean;
}) {
  const copyApiKeyToClickBoard = async (key: string) => {
    try {
      await navigator.clipboard.writeText(key);
      toast({
        title: "Copied!",
        variant: "success",
      });
    } catch (error) {
      toast({
        title: "Copy to clipboard failed",
        description:
          (error as Error).message ||
          "You're likely using an old browser that doesn't support this feature.",
        variant: "destructive",
      });
    }
  };
  return (
    <tr key={name} title={name}>
      <td className="text-left text-sm">{text}</td>
      <td className="text-right text-xs">{formatFileName(name)}</td>
      {showCopy && (
        <td
          className="w-1 text-right text-xs"
          onClick={() => copyApiKeyToClickBoard(name)}
        >
          <LucideCopy className="hover-view cursor-pointer" size={10} />
        </td>
      )}
    </tr>
  );
}

export function ProjectFiles({ data, title, field, showCopy = false }: iProps) {
  const entities = data.reduce(
    (prev: any[], curr) => [...prev, ...curr[field]],
    []
  );

  return (
    <div>
      <h1 className="text-center">{title}</h1>
      <table className="entities-table">
        <tbody>
          {entities.map((entity: any) => (
            <TableRow
              key={entity.name}
              name={entity.name}
              text={entity.text}
              showCopy={showCopy}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}
