"use client";

import React from "react";
import { Menu, X } from "lucide-react";

import Image from "next/image";

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
        {isOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
      </Button>
      {isOpen && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
          <div className="fixed inset-y-0 left-0 z-50 h-full w-full max-w-xs main-bg bg-white p-6 shadow-lg">
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-4 top-4"
              onClick={() => setIsOpen(false)}
            >
              <X className="h-5 w-5" />
              <span className="sr-only">Close</span>
            </Button>
            <div className="mt-8">
              <NavItems onItemClick={() => setIsOpen(false)} />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export function PublicMobileHeader() {
  return (
    <div className="md:hidden flex align-middle">
      <header className="flex items-center justify-between align-middle w-full p-4">
        <a href="/">
          <Image
            src={"/logo/white-icon.png"}
            alt="Logo"
            width={80}
            height={56}
            className="logo-icon-white"
          />
        </a>
        <OverlayNav />
      </header>
    </div>
  );
}
