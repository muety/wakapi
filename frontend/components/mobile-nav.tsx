"use client";

import {
  FolderGit2,
  Goal,
  LayoutDashboardIcon,
  Receipt,
  Trophy,
  UsersIcon,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/utils";

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

export function MobileNav() {
  return (
    <>
      <ul className="sticky flex w-full flex-col align-middle">
        {MAIN_MENU_ITEMS.map((menu, index) => (
          <SideNavMenuItem
            title={menu.title}
            icon={menu.icon}
            href={menu.href}
            key={index}
          />
        ))}
        <SideNavSimpleMenuItem title="FAQ" href="/faq" />
        <SideNavSimpleMenuItem title="About" href="/about" />
        <SideNavSimpleMenuItem title="Plugin Status" href="/plugins/status" />
        <SideNavSimpleMenuItem
          title="Logout"
          href="/api/session?action=logout"
        />
      </ul>
    </>
  );
}
