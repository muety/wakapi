"use client";

import { FrownIcon } from "lucide-react";
import { useRouter } from "next/navigation";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const router = useRouter();

  console.log({ error });

  return (
    <main
      className="flex h-full flex-col items-center justify-center gap-4"
      style={{ minHeight: "70vh" }}
    >
      <FrownIcon />
      <h2 className="text-center">Unexpected Error OccurreD!</h2>
      <button
        className="rounded-md px-4 py-2 text-sm text-white transition-colors"
        onClick={() => {
          reset();
          router.refresh();
        }}
      >
        Retry
      </button>
    </main>
  );
}
