"use client";

import { LucidePen, LucidePenOff, LucideSave } from "lucide-react";
import React from "react";

import { toast } from "@/components/ui/use-toast";
import { NEXT_PUBLIC_API_URL } from "@/lib/constants/config";
import useSession from "@/lib/session/use-session";

import { Icons } from "./icons";
import { Button } from "./ui/button";
import { Input } from "./ui/input";

/**
 * Wakatime Api key integration is one-way.
 * We only update the key for the integration and let the user
 * know whether or not they have one set.
 */

const RESOURCE_URL = `${NEXT_PUBLIC_API_URL}/api/settings`;

export function WakatimeIntegration({ token }: { token: string }) {
  const [copied, setCopied] = React.useState(false);
  const [saving, setSaving] = React.useState(false);
  const [showEditable, setShowEditable] = React.useState(false);
  const [apiKey, setApiKey] = React.useState("");

  const { modifySession } = useSession();

  React.useEffect(() => {
    if (copied) {
      setTimeout(() => {
        setCopied(false);
      }, 6000);
    }
  }, [copied]);

  const saveApiKey = async () => {
    try {
      setSaving(true);

      const payload = {
        action: "toggle_wakatime",
        api_url: "https://api.wakatime.com/api/v1",
        api_key: apiKey,
      };

      const response = await fetch(RESOURCE_URL, {
        method: "POST",
        body: JSON.stringify(payload),
        headers: {
          "content-type": "application/json",
          accept: "application/json",
          token: `${token}`,
        },
      });
      if (!response.ok) {
        throw new Error("Error setting wakatime api key");
      }
      modifySession({ has_wakatime_integration: true }); // allow to fail silently
      setShowEditable(false);
      toast({
        title: "Wakatime api key saved",
        variant: "success",
      });
    } catch (error) {
      toast({
        title: "Failed to save wakatime api key",
        description:
          "Your wakatime api key could not be saved at the moment. Try again later.",
        variant: "destructive",
      });
      console.error(error);
    } finally {
      setTimeout(() => {
        setSaving(false);
        setShowEditable(false);
      }, 6000);
    }
  };

  const dummyApiKey = "*******************";

  return (
    <div className="flex gap-2">
      {showEditable ? (
        <Input
          className="h-9 py-0"
          placeholder="Wakatime api key"
          defaultValue={apiKey}
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
        />
      ) : (
        <Input
          className="h-9 py-0"
          placeholder={dummyApiKey}
          disabled
          defaultValue={dummyApiKey}
          value={dummyApiKey}
        />
      )}
      <Button
        size={"sm"}
        variant="outline"
        onClick={() => setShowEditable(!showEditable)}
      >
        {showEditable ? (
          <LucidePenOff size="15" />
        ) : (
          <LucidePen color="gray" size="15" />
        )}
      </Button>
      {showEditable && (
        <Button
          size={"sm"}
          variant="outline"
          disabled={apiKey.length === 0}
          onClick={saveApiKey}
        >
          {saving ? (
            <Icons.spinner className="mr-2 size-4 animate-spin" />
          ) : (
            <LucideSave size="15" />
          )}
        </Button>
      )}
    </div>
  );
}
