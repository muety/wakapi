"use client";

import { Loader2 } from "lucide-react";
import { useState } from "react";

import { updatePreference } from "@/actions/update-preferences";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { toast } from "@/components/ui/use-toast";
import { UserProfile } from "@/lib/types";

export function UserPreferences({ user }: { user: UserProfile }) {
  const [hireable, setHireableBadge] = useState(user?.hireable);
  const [show_email_in_public, setDisplayEmail] = useState(
    user?.show_email_in_public
  );
  const [public_leaderboard, setDisplayCodeTime] = useState(
    user?.public_leaderboard
  );
  const [loading, setLoading] = useState<string | null>(null);

  const handleToggle = async (
    preference: "hireable" | "show_email_in_public" | "public_leaderboard",
    value: boolean
  ) => {
    try {
      setLoading(preference);
      const newValue = (await updatePreference(preference, value)) as boolean;
      switch (preference) {
        case "hireable":
          setHireableBadge(newValue);
          break;
        case "show_email_in_public":
          setDisplayEmail(newValue);
          break;
        case "public_leaderboard":
          setDisplayCodeTime(newValue);
          break;
      }
    } catch (error) {
      toast({
        title: "Error updating preference",
        description: (error as Error).message,
        variant: "destructive",
      });
    } finally {
      setLoading(null);
    }
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
            {loading === "hireable" && (
              <Loader2 className="size-4 animate-spin" />
            )}
            <Switch
              id="hireable-badge"
              checked={hireable}
              onCheckedChange={(value) => handleToggle("hireable", value)}
              disabled={loading === "hireable"}
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
            {loading === "show_email_in_public" && (
              <Loader2 className="size-4 animate-spin" />
            )}
            <Switch
              id="display-email"
              checked={show_email_in_public}
              onCheckedChange={(value) =>
                handleToggle("show_email_in_public", value)
              }
              disabled={loading === "show_email_in_public"}
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
            {loading === "public_leaderboard" && (
              <Loader2 className="size-4 animate-spin" />
            )}
            <Switch
              id="display-code-time"
              checked={public_leaderboard}
              onCheckedChange={(value) =>
                handleToggle("public_leaderboard", value)
              }
              disabled={loading === "public_leaderboard"}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
