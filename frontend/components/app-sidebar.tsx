import {
  FolderGit2,
  Goal,
  Info,
  LayoutDashboardIcon,
  Quote,
  Receipt,
  SquareActivity,
  Trophy,
  UsersIcon,
} from "lucide-react";

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

import { HeroBrand } from "./hero-brand";

export const MAIN_MENU_ITEMS = [
  {
    title: "Dashboard",
    url: "/dashboard",
    icon: LayoutDashboardIcon,
  },
  {
    title: "Projects",
    url: "/projects",
    icon: FolderGit2,
  },
  {
    title: "Goals",
    url: "/goals",
    icon: Goal,
  },
  {
    title: "Clients",
    url: "/clients",
    icon: UsersIcon,
  },
  {
    title: "Invoices",
    url: "/invoices",
    icon: Receipt,
  },
  {
    title: "Leaderboards",
    url: "/dashboard/leaderboards",
    icon: Trophy,
  },
];

const SIMPLE_MENU_ITEMS = [
  {
    title: "FAQ",
    url: "/faq",
    icon: Quote,
  },
  {
    title: "About",
    url: "/about",
    icon: Info,
  },
  {
    title: "Plugin Status",
    url: "/plugins/status",
    icon: SquareActivity,
  },
];

export function AppSidebar() {
  return (
    <Sidebar collapsible="icon">
      <SidebarContent>
        <div className="brand-logo py-6 pl-2">
          <HeroBrand
            imgHeight={25}
            imgWidth={25}
            fontSize="32px"
            lineHeight="25px"
            inDashboard={true}
          />
        </div>
        <div className="flex h-screen flex-col justify-between">
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                {MAIN_MENU_ITEMS.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton asChild>
                      <a href={item.url}>
                        <item.icon />
                        <span>{item.title}</span>
                      </a>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>

          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                {SIMPLE_MENU_ITEMS.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton asChild>
                      <a href={item.url}>
                        <item.icon />
                        <span>{item.title}</span>
                      </a>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </div>
      </SidebarContent>
    </Sidebar>
  );
}
