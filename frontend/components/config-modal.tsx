"use client";

import { Check, Copy } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface ConfigModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  apiKey: string;
  apiUrl: string;
}

export function ConfigModal({
  open,
  onOpenChange,
  apiKey,
  apiUrl,
}: ConfigModalProps) {
  const [copied, setCopied] = useState(false);

  const configContent = `[settings]
api_url = ${apiUrl}
api_key = ${apiKey}`;

  const copyToClipboard = () => {
    navigator.clipboard.writeText(configContent);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md w-[calc(100%-2rem)] p-4 sm:p-6">
        <DialogHeader>
          <DialogTitle className="text-xl">Sample Configuration</DialogTitle>
          <DialogDescription className="text-sm sm:text-base">
            Copy and paste this into your <code>.wakatime.cfg</code> file
          </DialogDescription>
        </DialogHeader>
        <div className="relative mt-2">
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
          After updating your configuration, restart your editor for the changes
          to take effect.
        </div>
      </DialogContent>
    </Dialog>
  );
}
