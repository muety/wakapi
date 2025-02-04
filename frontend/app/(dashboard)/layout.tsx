import { cookies } from "next/headers";
import React from "react";

import { AppSidebar } from "@/components/app-sidebar";
import { DashboardHeader } from "@/components/dashboard-header";
import { MobileHeader } from "@/components/mobile-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";

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
      <SidebarInset className="shadow-lg border border-[##dddddd]">
        <DashboardHeader />
        <MobileHeader />
        <main className="min-h-full px-5" style={{ minHeight: "50vh" }}>
          {children}
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
