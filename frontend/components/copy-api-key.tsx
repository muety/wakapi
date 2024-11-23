"use client";

import React from "react";

import {
  LucideCopy,
  LucideCopyCheck,
  LucideEye,
  LucideEyeOff,
  LucideRefreshCcw,
} from "lucide-react";
import { Icons } from "./icons";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import { toast } from "./ui/use-toast";
import { Confirm } from "./ui/confirm";
import { copyApiKeyToClickBoard } from "@/lib/utils/ui";
import { NEXT_PUBLIC_API_URL } from "@/lib/constants/config";

export function ApiKeyCopier({ token }: { token: string }) {
  const [copied, setCopied] = React.useState(false);
  const [masked, setMasked] = React.useState(true);
  const [loading, setLoading] = React.useState(false);
  const [apiKey, setApiKey] = React.useState("");

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

  const refreshApiKey = async () => {
    try {
      setLoading(true);
      const resourceUrl = `${NEXT_PUBLIC_API_URL}/api/auth/api-key/refresh`;
      const response = await fetch(resourceUrl, {
        method: "POST",
        headers: {
          accept: "application/json",
          "content-type": "application/json",
          token: `${token}`,
        },
      });

      if (!response.ok) {
        toast({
          title: "Failed to refresh api key",
          variant: "destructive",
        });
      } else {
        const data = await response.json();
        setMasked(false);
        setApiKey(data.apiKey);
        toast({
          title: "Api key refreshed",
          description: `Api key refreshed. Copy and use this api key to use in your IDE.`,
          variant: "success",
        });
      }
    } finally {
      setLoading(false);
    }
  };

  const showApiKey = () => {
    if (!masked) {
      return setMasked(!masked);
    }
    if (apiKey) {
      setMasked(false);
    } else {
      getApiKey();
    }
  };

  React.useEffect(() => {
    if (copied) {
      setTimeout(() => {
        setCopied(false);
      }, 6000);
    }
  }, [copied]);

  const maskedKey = "*".repeat(24);

  return (
    <div className="flex gap-2">
      <Input
        className="py-0 h-9"
        placeholder="Api Key"
        disabled
        defaultValue={maskedKey}
        value={masked ? maskedKey : apiKey}
      />
      <Button size={"sm"} variant="outline" onClick={showApiKey}>
        {loading && <Icons.spinner className="mr-2 size-4 animate-spin" />}
        {!masked ? (
          <LucideEyeOff size="15" />
        ) : (
          <LucideEye color="gray" size="15" />
        )}
      </Button>
      <Confirm
        title="Refresh Api Key?"
        description="This will invalidate your current API key and generate a new one."
        onConfirm={() => refreshApiKey()}
        loading={loading}
      >
        <Button
          size={"sm"}
          variant="outline"
          onClick={() => setMasked(!masked)}
        >
          <LucideRefreshCcw size="15" />
        </Button>
      </Confirm>

      <Button
        size={"sm"}
        variant="outline"
        onClick={() =>
          copyApiKeyToClickBoard(apiKey).then(() => setCopied(true))
        }
      >
        {!copied ? (
          <LucideCopy size="15" />
        ) : (
          <LucideCopyCheck size="15" color="green" />
        )}
      </Button>
    </div>
  );
}
