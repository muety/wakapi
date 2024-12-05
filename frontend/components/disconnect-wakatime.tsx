"use client";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { LucideUnplug } from "lucide-react";
import { Confirm } from "@/components/ui/confirm";

export function DisconnectWakatime() {
  const disconnectWakatimeIntegration = () => {
    console.log("");
  };

  const disconnecting = false;
  return (
    <Confirm
      title="Disconnect wakatime"
      description={`Your heartbeats will no longer be relayed to wakatime. Wakatime will no longer receive your plugin activity. Your encrypted wakatime api key will be deleted from our database.`}
      onConfirm={() => disconnectWakatimeIntegration()}
      loading={disconnecting}
      continueClassName="bg-red-500 hover:bg-red-600 text-gray-200"
      titleClassName="text-red-500"
    >
      <Button
        className={cn(
          "bg-red-500 hover:bg-red-600 text-gray-200 px-3 text-sm rounded-md"
        )}
        size="sm"
      >
        <LucideUnplug className="size-4 cursor-pointer text-gray-200 " />
        Disconnect
      </Button>
    </Confirm>
  );
}
