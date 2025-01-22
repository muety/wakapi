import {
  FolderGit2,
  Goal,
  Info,
  LayoutDashboardIcon,
  Quote,
  Receipt,
  SquareActivity,
  TrendingUpIcon,
  Trophy,
  UsersIcon,
} from "lucide-react";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

import { AppSidebarLogo } from "./app-sidebar-logo";
import { NavUser } from "./nav-user";

const SIMPLE_MENU_ITEMS = [
  {
    title: "Leaderboards",
    url: "/dashboard/leaderboards",
    icon: Trophy,
  },
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

export const MAIN_MENU_ITEMS = [
  {
    group: "Dashboard",
    children: [
      {
        title: "Dashboard",
        url: "/dashboard",
        icon: LayoutDashboardIcon,
      },
      {
        title: "Analytics",
        url: "/analytics",
        icon: TrendingUpIcon,
      },
    ],
  },
  {
    group: "Projects",
    children: [
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
    ],
  },
  {
    group: "Freelance",
    children: [
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
    ],
  },
  {
    group: "Miscellaneous",
    children: SIMPLE_MENU_ITEMS,
  },
];

export function AppSidebar() {
  return (
    <Sidebar collapsible="icon" variant="inset">
      <SidebarContent>
        <AppSidebarLogo />
        <div className="flex h-screen flex-col">
          {MAIN_MENU_ITEMS.map((item) => (
            <SidebarGroup key={item.group}>
              <SidebarGroupLabel>{item.group}</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  {item.children.map((item) => (
                    <SidebarMenuItem key={item.title}>
                      <SidebarMenuButton
                        asChild
                        // isActive={sidebarActive(item)}
                        // @todo: move this to client component and implement isActive
                      >
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
          ))}
        </div>
      </SidebarContent>
      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
    </Sidebar>
  );
}
