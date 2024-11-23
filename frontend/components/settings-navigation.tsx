"use client";

import { usePathname } from "next/navigation";
import { SettingItem } from "./settings-menu-item";

export function SettingsNavigation() {
  return (
    <nav className="grid gap-0 text-sm" x-chunk="dashboard-04-chunk-0">
      <SettingItem key="General" href="/settings" title="General" />
      <SettingItem
        key="Preferences"
        href="/settings/preferences"
        title="Preferences"
      />
    </nav>
  );
}
