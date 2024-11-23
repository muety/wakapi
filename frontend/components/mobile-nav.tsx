"use client";
import Link from "next/link";
import { Goal, Receipt, Trophy, FolderGit2, Gauge } from "lucide-react";
import { HeroBrand } from "./hero-brand";
import { cn } from "@/lib/utils";
import { usePathname } from "next/navigation";
import {
  MAIN_MENU_ITEMS,
  SideNavMenuItem,
  SideNavSimpleMenuItem,
} from "./side-nav";

export function MobileNav() {
  return (
    <>
      <ul className="sticky w-full flex flex-col align-middle">
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
