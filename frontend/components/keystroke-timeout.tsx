"use client";

import { LucideSave } from "lucide-react";
import React from "react";

import { updatePreference } from "@/actions/update-preferences";

import { Icons } from "./icons";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { toast } from "./ui/use-toast";

export function KeystrokeTimeout({
  initialValue,
}: {
  token: string;
  initialValue?: number;
}) {
  const [loading, setLoading] = React.useState(false);
  const [keyStrokeTimeout, setKeyStrokeTimeout] = React.useState<number>(
    initialValue || 0
  );

  const saveKeystrokeTimeout = async () => {
    if (!keyStrokeTimeout) {
      return;
    }
    try {
      setLoading(true);
      await updatePreference("heartbeats_timeout_sec", keyStrokeTimeout);
      toast({
        title: "Saved keystroke timeout",
        description: `Keystroke timeout saved. Future summaries will now use this to compute time spent.`,
        variant: "success",
      });
    } catch (error) {
      toast({
        title: "Error updating keystroke timeout",
        description: (error as Error).message,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex gap-2">
      <Input
        type="number"
        className="h-9 py-0"
        placeholder="Keystroke Timeout"
        disabled={loading}
        value={keyStrokeTimeout}
        onChange={(e) => setKeyStrokeTimeout(+e.target.value)}
      />
      <Button
        variant="outline"
        disabled={loading}
        onClick={saveKeystrokeTimeout}
      >
        {loading ? (
          <Icons.spinner className="mr-2 size-4 animate-spin" />
        ) : (
          <LucideSave size="15" />
        )}
      </Button>
    </div>
  );
}
