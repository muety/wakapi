import React, { Suspense } from "react";

import Loading from "./loading";

export default async function Layout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <Suspense fallback={<Loading />}>
      <main className="min-h-full md:px-5 w-full" style={{ minHeight: "50vh" }}>
        {children}
      </main>
    </Suspense>
  );
}
