import { CurrentWorkTime } from "./current-work-time";
import { SidebarTrigger } from "./ui/sidebar";
import { UserDropdownMenu } from "./user-dropdown-menu";

export function DashboardHeader() {
  return (
    // sticky top-0 sm-static sm:h-auto
    <header className="dashboard-header md:border-1 h-18 top-0 hidden w-full items-center gap-3 px-2 py-4 pr-5 md:flex md:justify-between">
      <SidebarTrigger />
      <div className="flex gap-4 align-middle">
        <CurrentWorkTime />
        <UserDropdownMenu />
      </div>
    </header>
  );
}
