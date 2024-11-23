"use client";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import Image from "next/image";
import React from "react";
import { Button } from "./ui/button";
import { SessionUser } from "@/lib/session/options";
import { useSessionUser } from "@/lib/session/use-session";

// interface UserAccountNavProps extends React.HTMLAttributes<HTMLDivElement> {
//   user: Pick<SessionUser, "id" | "avatar" | "email">;
// }

export function UserDropdownMenu() {
  const { data: user, isLoading } = useSessionUser();
  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            size="icon"
            className="overflow-hidden rounded-full"
          >
            <Image
              src={user?.avatar || ""}
              width={36}
              height={36}
              alt="Avatar"
              className="overflow-hidden"
            />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" style={{ zIndex: 100 }}>
          <DropdownMenuItem className="cursor-pointer">
            <a href="/settings" className="cursor-pointer">
              Settings
            </a>
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <a href="/api/session?action=logout" className="cursor-pointer">
              Logout
            </a>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
