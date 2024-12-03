import Image from "next/image";
import Link from "next/link";
import React from "react";

import { HeroBrand } from "@/components/hero-brand";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface MarketingLayoutProps {
  children: React.ReactNode;
}

export function ClassicHeader() {
  return (
    <header className="w-100 w-full pt-8">
      <div className="flex h-20 w-full items-center justify-between">
        <HeroBrand fontSize="42px" logoType="white" />
        <div className="flex items-center gap-6">
          <nav>
            <Link href="/leaderboards" className="hero-link">
              Leaderboard
            </Link>
          </nav>
          <nav>
            <Link
              href="/auth/signin"
              className={cn(
                buttonVariants({ variant: "outline", size: "sm" }),
                "px-4"
              )}
              style={{ borderRadius: "5px" }}
            >
              Login
            </Link>
          </nav>
        </div>
      </div>
    </header>
  );
}

function Header() {
  return (
    <header className="sticky top-8 z-50 mb-20 mt-8 justify-center rounded-lg px-2 md:flex md:px-4">
      <nav className="z-20 flex h-[50px] items-center rounded-full border border-border bg-[#FFFFFF] bg-opacity-70 px-4 backdrop-blur-xl dark:bg-[#121212]">
        <a href="/">
          <Image
            src={"/logo/white-icon.png"}
            alt="Logo"
            width={80}
            height={56}
            className="logo-icon-white"
          />
        </a>
        <ul className="mx-3 hidden space-x-2 text-sm font-medium md:flex">
          <a
            href="/pricing"
            className="inline-flex h-8 items-center justify-center px-3 py-2 text-sm font-medium text-secondary-foreground transition-opacity duration-200 hover:opacity-70"
          >
            Pricing
          </a>
          <a
            href="/faqs"
            className="inline-flex h-8 items-center justify-center px-4 py-2 text-sm font-medium text-secondary-foreground transition-opacity duration-200 hover:opacity-70"
          >
            FAQ
          </a>
          <a
            href="/plugins"
            className="inline-flex h-8 items-center justify-center px-3 py-2 text-sm font-medium text-secondary-foreground transition-opacity duration-200 hover:opacity-70"
          >
            Plugins
          </a>
          <a
            href="/leaderboards"
            className="inline-flex h-8 items-center justify-center px-3 py-2 text-sm font-medium text-secondary-foreground transition-opacity duration-200 hover:opacity-70"
          >
            Leaderboard
          </a>
        </ul>
        <a
          className="hidden border-l border-border pl-4 pr-2 text-sm font-medium md:block"
          href="/auth/signin"
        >
          Sign in
        </a>
      </nav>
    </header>
  );
}

export default async function Page({ children }: MarketingLayoutProps) {
  return (
    <div className="flex min-h-screen flex-col">
      <Header />
      <main className="flex-1">{children}</main>
      <footer>
        <h5 className="pointer-events-none mx-12 text-center text-[500px] leading-none text-[#161616]">
          wakana
        </h5>
      </footer>
    </div>
  );
}
