import React from "react";
import Link from "next/link";

import { Button } from "./ui/button";
import { SessionData } from "@/lib/session/options";
import { CurrentWorkTime } from "./current-work-time";
import { UserDropdownMenu } from "./user-dropdown-menu";
import { SidebarTrigger } from "./ui/sidebar";

export function DashboardHeader() {
  return (
    // sticky top-0 sm-static sm:h-auto
    <header className="dashboard-header top-0 hidden md:flex py-4 md:justify-between items-center gap-3 md:border-1 px-2 pr-5 w-full h-18">
      <SidebarTrigger />
      <div className="flex gap-4 align-middle">
        <Button className="p-1 py-0 m-1 text-sm h-8">
          <Link href="/dashboard">Dashboard</Link>
        </Button>
        <CurrentWorkTime />
        <UserDropdownMenu />
      </div>
    </header>
  );
}
