import Link from "next/link";

export default function NotFound() {
  return (
    <main
      className="flex h-full flex-col items-center justify-center gap-2"
      style={{ minHeight: "80vh" }}
    >
      <h2 className="text-xl font-semibold">404 Not Found</h2>
      <hr />
      <p>Could not find the requested invoice.</p>
      <Link
        href="/invoices"
        className="mt-4 rounded-md px-4 py-2 text-sm text-white transition-colors"
      >
        Back to Invoices
      </Link>
    </main>
  );
}
