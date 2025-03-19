import Image from "next/image";
import Link from "next/link";
import React from "react";

import { FadeOnView } from "@/components/fade-on-view";

import PublicFooter from "./sections/components/public-footer";
import { PublicMobileHeader } from "./sections/public-mobile-header";

interface MarketingLayoutProps {
  children: React.ReactNode;
}

function Header() {
  return (
    <header className="sticky top-8 z-50 mb-20 mt-8 hidden justify-center rounded-lg px-2 md:flex md:px-4">
      <nav className="z-20 flex h-[50px] items-center rounded-full border border-border bg-[#FFFFFF] bg-opacity-70 px-4 backdrop-blur-xl dark:bg-[#121212]">
        <Link href="/">
          <Image
            src={"/white-icon.svg"}
            alt="Logo"
            width={80}
            height={56}
            className="logo-icon-white"
          />
        </Link>
        <ul className="mx-3 space-x-2 text-sm font-medium md:flex">
          <a
            href="/setup"
            className="inline-flex h-8 items-center justify-center px-3 py-2 text-sm font-medium text-secondary-foreground transition-opacity duration-200 hover:opacity-70"
          >
            Installation
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
          href="/login"
        >
          Sign in
        </a>
      </nav>
    </header>
  );
}

export default async function Page({ children }: MarketingLayoutProps) {
  return (
    <div className="flex flex-col min-h-screen">
      <Header />
      <PublicMobileHeader />
      <main className="m-auto md:mx-14 flex flex-1 flex-col px-12 md:px-14 align-middle">
        <FadeOnView>{children}</FadeOnView>
      </main>
      <PublicFooter />
    </div>
  );
}
