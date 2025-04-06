"use client";

import { Check, Copy } from "lucide-react";
import { useEffect, useState } from "react";
import React from "react";

import { Button } from "@/components/ui/button";
import { NEXT_PUBLIC_API_URL } from "@/lib/constants/config";
import useSession from "@/lib/session/use-session";

import { Spinner } from "./spinner/spinner";
import { toast } from "./ui/use-toast";

interface ConfigDisplayProps {
  className?: string;
}

const API_URL = `${NEXT_PUBLIC_API_URL}/api`;

export function Installation({ className = "" }: ConfigDisplayProps) {
  const [copied, setCopied] = useState(false);
  const [loading, setLoading] = useState(false);
  const [apiKey, setApiKey] = useState("2df5863c-86ec-590b-a9d3-1755c28c39da");

  const { session } = useSession();

  const getApiKey = React.useCallback(async () => {
    try {
      const token = session?.token;
      if (!token) {
        return;
      }
      setLoading(true);
      const resourceUrl = `${NEXT_PUBLIC_API_URL}/api/v1/auth/api-key`;
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
      }
    } finally {
      setLoading(false);
    }
  }, [session]);

  const isLoggedIn = session?.isLoggedIn;

  useEffect(() => {
    if (isLoggedIn) {
      console.log("isLoggedIn", isLoggedIn);
      getApiKey();
    }
  }, [isLoggedIn, getApiKey]);

  const configContent = React.useMemo(() => {
    return `[settings]
api_url = ${API_URL}
api_key = ${isLoggedIn ? apiKey : "## replace this with your api key when you login"}
`;
  }, [apiKey, isLoggedIn]);

  const copyToClipboard = () => {
    navigator.clipboard.writeText(configContent);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className={`w-full ${className}`}>
      <div className="mb-4">
        {isLoggedIn ? (
          <h2 className="text-xl font-semibold">Your ~/.wakatime.cfg</h2>
        ) : (
          <h2 className="text-xl font-semibold">Sample Configuration</h2>
        )}
        <p className="text-sm text-muted-foreground">
          Copy and paste this into your <code>~/.wakatime.cfg</code> file
        </p>
      </div>

      <div className="relative mt-2">
        {loading && <Spinner />}
        <pre className="bg-muted p-3 sm:p-4 rounded-md overflow-x-auto text-sm whitespace-pre-wrap break-words">
          {configContent}
        </pre>
        <Button
          size="icon"
          variant="ghost"
          className="absolute top-2 right-2 h-9 w-9 sm:h-8 sm:w-8"
          onClick={copyToClipboard}
        >
          {copied ? (
            <Check className="h-5 w-5 sm:h-4 sm:w-4" />
          ) : (
            <Copy className="h-5 w-5 sm:h-4 sm:w-4" />
          )}
          <span className="sr-only">Copy</span>
        </Button>
      </div>

      <div className="text-sm text-muted-foreground mt-4 sm:mt-2">
        After updating your configuration, restart/reload your editor, type
        something and check your dashboard
      </div>
    </div>
  );
}
