"use client";

import {
  LucideCopy,
  LucideCopyCheck,
  LucideEye,
  LucideEyeOff,
  LucideRefreshCcw,
  LucideSave,
} from "lucide-react";
import React from "react";

import { NEXT_PUBLIC_API_URL } from "@/lib/constants/config";

import { Icons } from "./icons";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { toast } from "./ui/use-toast";

export function KeystrokeTimeout({ token }: { token: string }) {
  const [masked, setMasked] = React.useState(true);
  const [loading, setLoading] = React.useState(false);
  const [apiKey, setApiKey] = React.useState("");
  const [keyStrokeTimeout, setKeyStrokeTimeout] = React.useState("");

  const getApiKey = async () => {
    try {
      setLoading(true);
      const resourceUrl = `${NEXT_PUBLIC_API_URL}/api/auth/api-key`;
      const response = await fetch(resourceUrl, {
        method: "GET",
        headers: {
          accept: "application/json",
          "content-type": "application/json",
          token: `${token}`,
        },
      });

      if (!response.ok) {
        toast({
          title: "Failed to fetch api key",
          variant: "destructive",
        });
      } else {
        const data = await response.json();
        setApiKey(data.apiKey);
        setMasked(false);
      }
    } finally {
      setLoading(false);
    }
  };

  const saveKeystrokeTimeout = async () => {
    try {
      setLoading(true);
      const resourceUrl = `${NEXT_PUBLIC_API_URL}/api/settings`;
      const response = await fetch(resourceUrl, {
        method: "POST",
        body: JSON.stringify({
          action: "set_keystroke_timeout",
          keystroke_timeout: keyStrokeTimeout,
        }),
        headers: {
          accept: "application/json",
          "content-type": "application/json",
          token: `${token}`,
        },
      });

      if (!response.ok) {
        toast({
          title: "Failed to update keystroke timeout",
          variant: "destructive",
        });
      } else {
        const data = await response.json();
        setMasked(false);
        setApiKey(data.apiKey);
        toast({
          title: "Saved keystroke timeout",
          description: `Keystroke timeout saved. We've started recomputing your stats.`,
          variant: "success",
        });
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex gap-2">
      <Input
        className="h-9 py-0"
        placeholder="Keystroke Timeout"
        disabled={loading}
        value={keyStrokeTimeout}
        onChange={(e) => setKeyStrokeTimeout(e.target.value)}
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
