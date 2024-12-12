import { cookies } from "next/headers";
import React from "react";

import { AppSidebar } from "@/components/app-sidebar";
import { DashboardHeader } from "@/components/dashboard-header";
import { MobileHeader } from "@/components/mobile-header";
import { SidebarProvider } from "@/components/ui/sidebar";

export default async function Layout({
  children,
}: {
  children: React.ReactNode;
}) {
  const cookieStore = cookies();
  const defaultOpen = cookieStore.get("sidebar:state")?.value === "true";

  return (
    <SidebarProvider defaultOpen={defaultOpen}>
      <AppSidebar />
      <main className="w-full">
        <DashboardHeader />
        <MobileHeader />
        <main className="min-h-full px-5" style={{ minHeight: "50vh" }}>
          {children}
        </main>
      </main>
    </SidebarProvider>
  );
}
