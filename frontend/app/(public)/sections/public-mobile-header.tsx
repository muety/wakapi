"use client";

import { Menu, X } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import React from "react";

import { Button } from "@/components/ui/button";

import { NavItems } from "./components/nav-items";

export function OverlayNav() {
  const [isOpen, setIsOpen] = React.useState(false);

  return (
    <div className="relative md:hidden">
      <Button
        variant="outline"
        size="icon"
        onClick={() => setIsOpen(!isOpen)}
        aria-label="Toggle menu"
      >
        {isOpen ? <X className="size-5" /> : <Menu className="size-5" />}
      </Button>
      {isOpen && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
          <div className="main-bg fixed inset-y-0 left-0 z-50 size-full max-w-xs bg-white p-6 shadow-lg">
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-4 top-4"
              onClick={() => setIsOpen(false)}
            >
              <X className="size-5" />
              <span className="sr-only">Close</span>
            </Button>
            <div className="mt-8">
              <NavItems />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export function PublicMobileHeader() {
  return (
    <div className="flex align-middle md:hidden">
      <header className="flex w-full items-center justify-between p-4 align-middle">
        <Link href="/">
          <Image
            src={"/white-icon.svg"}
            alt="Logo"
            width={80}
            height={56}
            className="logo-icon-white"
          />
        </Link>
        <OverlayNav />
      </header>
    </div>
  );
}
