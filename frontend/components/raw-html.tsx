import { preserveNewLine } from "@/lib/utils";

export const RawHTML = ({
  source,
  fallback = "",
}: {
  source: string;
  fallback?: string;
}) => {
  return (
    <div
      className="prose"
      dangerouslySetInnerHTML={{
        __html: preserveNewLine(source, fallback),
      }}
    />
  );
};
