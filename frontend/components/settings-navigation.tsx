"use client";

import { SettingItem } from "./settings-menu-item";

export function SettingsNavigation() {
  return (
    <nav className="grid gap-0 text-sm">
      <SettingItem key="General" href="/settings" title="General" />
      <SettingItem key="My Profile" href="/settings/profile" title="Profile" />
      <SettingItem
        key="Preferences"
        href="/settings/preferences"
        title="Preferences"
      />
    </nav>
  );
}
