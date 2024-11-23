import { cookies } from "next/headers";
import { AppSidebar } from "@/components/app-sidebar";
import { SidebarProvider } from "@/components/ui/sidebar";
import { MobileHeader } from "@/components/mobile-header";
import { DashboardHeader } from "@/components/dashboard-header";

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
        <main
          className="min-h-full min-h-100 px-5"
          style={{ minHeight: "50vh" }}
        >
          {children}
        </main>
      </main>
    </SidebarProvider>
  );
}
