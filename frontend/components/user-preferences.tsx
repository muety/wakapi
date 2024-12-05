"use client";

import { useState } from "react";
import { Loader2 } from "lucide-react";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { updatePreference } from "@/actions/update-preferences";

export function UserPreferences() {
  const [hireableBadge, setHireableBadge] = useState(false);
  const [displayEmail, setDisplayEmail] = useState(false);
  const [displayCodeTime, setDisplayCodeTime] = useState(false);
  const [loading, setLoading] = useState<string | null>(null);

  const handleToggle = async (
    preference: "hireableBadge" | "displayEmail" | "displayCodeTime",
    value: boolean
  ) => {
    setLoading(preference);
    const newValue = await updatePreference(preference, value);
    switch (preference) {
      case "hireableBadge":
        setHireableBadge(newValue);
        break;
      case "displayEmail":
        setDisplayEmail(newValue);
        break;
      case "displayCodeTime":
        setDisplayCodeTime(newValue);
        break;
    }
    setLoading(null);
  };

  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="hireable-badge">Hireable Badge</Label>
            <p className="text-sm text-muted-foreground">
              Display a badge indicating you're open for hire
            </p>
          </div>
          <div className="flex items-center space-x-2">
            {loading === "hireableBadge" && (
              <Loader2 className="h-4 w-4 animate-spin" />
            )}
            <Switch
              id="hireable-badge"
              checked={hireableBadge}
              onCheckedChange={(value) => handleToggle("hireableBadge", value)}
              disabled={loading === "hireableBadge"}
            />
          </div>
        </div>
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="display-email">Display Email Publicly</Label>
            <p className="text-sm text-muted-foreground">
              Show your email address on your public profile
            </p>
          </div>
          <div className="flex items-center space-x-2">
            {loading === "displayEmail" && (
              <Loader2 className="h-4 w-4 animate-spin" />
            )}
            <Switch
              id="display-email"
              checked={displayEmail}
              onCheckedChange={(value) => handleToggle("displayEmail", value)}
              disabled={loading === "displayEmail"}
            />
          </div>
        </div>
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <Label htmlFor="display-code-time">Public Leaderboard</Label>
            <p className="text-sm text-muted-foreground">
              You'll be shown on our public leaderboard
            </p>
          </div>
          <div className="flex items-center space-x-2">
            {loading === "displayCodeTime" && (
              <Loader2 className="h-4 w-4 animate-spin" />
            )}
            <Switch
              id="display-code-time"
              checked={displayCodeTime}
              onCheckedChange={(value) =>
                handleToggle("displayCodeTime", value)
              }
              disabled={loading === "displayCodeTime"}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
