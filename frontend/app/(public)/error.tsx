"use client";

import { FrownIcon } from "lucide-react";

export default function Error({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <main
      className="flex h-full flex-col items-center justify-center gap-4"
      style={{ minHeight: "70vh" }}
    >
      <FrownIcon />
      <h2 className="text-center">Unexpected Error Occurred!</h2>
      <button
        className="rounded-md px-4 py-2 text-sm text-white transition-colors"
        onClick={() => reset()}
      >
        Retry
      </button>
    </main>
  );
}
