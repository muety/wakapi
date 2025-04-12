"use client";

import {
  LucideCopy,
  LucideCopyCheck,
  LucideEye,
  LucideEyeOff,
  LucideRefreshCcw,
} from "lucide-react";
import React from "react";

import { getData, postData } from "@/actions/api";
import { copyApiKeyToClickBoard } from "@/lib/utils/ui";

import { Icons } from "./icons";
import { Button } from "./ui/button";
import { Confirm } from "./ui/confirm";
import { Input } from "./ui/input";
import { toast } from "./ui/use-toast";

export function ApiKeyCopier() {
  const [copied, setCopied] = React.useState(false);
  const [masked, setMasked] = React.useState(true);
  const [loading, setLoading] = React.useState(false);
  const [apiKey, setApiKey] = React.useState("");

  const getApiKey = React.useCallback(async () => {
    try {
      setLoading(true);
      const response = await getData<{ apiKey: string }>("/v1/auth/api-key");

      if (!response.success) {
        toast({
          title: "Failed to fetch api key",
          variant: "destructive",
        });
      } else {
        const data = response.data;
        setApiKey(data.apiKey);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  const refreshApiKey = async () => {
    try {
      setLoading(true);
      const resourceUrl = `/v1/auth/api-key/refresh`;
      const response = await postData<{ apiKey: string }>(resourceUrl, {});

      console.log("REFRESH RESPONSE");

      if (!response.success) {
        toast({
          title: "Failed to refresh api key",
          variant: "destructive",
        });
      } else {
        setMasked(false);
        setApiKey(response.data.apiKey);
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
    getApiKey();
  }, [getApiKey]);

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
        className="h-9 py-0"
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
