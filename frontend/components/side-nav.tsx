"use client";
import Link from "next/link";
import { cn } from "@/lib/utils";
import { usePathname } from "next/navigation";
import {
  Goal,
  Receipt,
  Trophy,
  FolderGit2,
  LayoutDashboardIcon,
  UsersIcon,
} from "lucide-react";
import { HeroBrand } from "./hero-brand";

export const MAIN_MENU_ITEMS = [
  {
    title: "Dashboard",
    href: "/dashboard",
    icon: <LayoutDashboardIcon size={16} />,
  },
  {
    title: "Projects",
    href: "/projects",
    icon: <FolderGit2 size={16} />,
  },
  {
    title: "Goals",
    href: "/goals",
    icon: <Goal size={16} />,
  },
  {
    title: "Clients",
    href: "/clients",
    icon: <UsersIcon size={16} />,
  },
  {
    title: "Invoices",
    href: "/invoices",
    icon: <Receipt size={16} />,
  },
  {
    title: "Leaderboards",
    href: "/leaderboards",
    icon: <Trophy size={16} />,
  },
];

export function SideNavMenuItem({ title, icon, href }: any) {
  const pathname = usePathname();
  const isActive = pathname.includes(href);

  return (
    <li>
      <Link
        href={href}
        className={cn("side-menu-item", isActive ? "active" : "")}
      >
        {icon}
        <span className="ml-2">{title}</span>
      </Link>
    </li>
  );
}

export function SideNavSimpleMenuItem({ title, href }: any) {
  const pathname = usePathname();
  const isActive = pathname.includes(href);
  return (
    <li>
      <Link
        href={href}
        className={cn("side-menu-item", isActive ? "active" : "")}
      >
        <span className="ml-2">{title}</span>
      </Link>
    </li>
  );
}

export function SideNav() {
  return (
    <aside className="hidden h-screen w-52 md:fixed md:flex md:flex-col side-nav z-52">
      <div className="side-nav-inner">
        <div className="top-side-nav">
          <div className="brand-logo pl-2 py-6">
            <HeroBrand
              imgHeight={25}
              imgWidth={25}
              fontSize="32px"
              lineHeight="25px"
            />
          </div>
          <div className="menu-top">
            <ul>
              {MAIN_MENU_ITEMS.map((menu, index) => (
                <SideNavMenuItem
                  title={menu.title}
                  icon={menu.icon}
                  href={menu.href}
                  key={index}
                />
              ))}
            </ul>
          </div>
        </div>
        <div className="menu-bottom">
          <ul>
            <SideNavSimpleMenuItem title="FAQ" href="/faq" />
            <SideNavSimpleMenuItem title="About" href="/about" />
            <SideNavSimpleMenuItem
              title="Plugin Status"
              href="/plugins/status"
            />
          </ul>
        </div>
      </div>
    </aside>
  );
}
