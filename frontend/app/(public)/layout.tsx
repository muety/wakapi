import Link from "next/link";
import Image from "next/image";

import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";
import { HeroBrand } from "@/components/hero-brand";

interface MarketingLayoutProps {
  children: React.ReactNode;
}

export function ClassicHeader() {
  return (
    <header className="pt-8 w-100 w-full">
      <div className="flex h-20 items-center justify-between w-full">
        <HeroBrand fontSize="42px" logoType="white" />
        <div className="flex gap-6 items-center">
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
    <header className="sticky mt-8 mb-20 top-8 z-50 px-2 md:px-4 md:flex justify-center rounded-lg">
      <nav className="border border-border px-4 flex items-center backdrop-filter backdrop-blur-xl bg-[#FFFFFF] dark:bg-[#121212] bg-opacity-70 h-[50px] z-20 rounded-full">
        <a href="/">
          <Image
            src={"/white-logo.svg"}
            alt="Logo"
            width={30}
            height={30}
            className="logo-icon-white"
          />
        </a>
        <ul className="space-x-2 font-medium text-sm hidden md:flex mx-3">
          <a
            href="/pricing"
            className="h-8 items-center justify-center text-sm font-medium px-3 py-2 inline-flex text-secondary-foreground transition-opacity hover:opacity-70 duration-200"
          >
            Pricing
          </a>
          <a
            href="/faqs"
            className="h-8 items-center justify-center text-sm font-medium px-4 py-2 inline-flex text-secondary-foreground transition-opacity hover:opacity-70 duration-200"
          >
            FAQ
          </a>
          <a
            href="/plugins"
            className="h-8 items-center justify-center text-sm font-medium px-3 py-2 inline-flex text-secondary-foreground transition-opacity hover:opacity-70 duration-200"
          >
            Plugins
          </a>
          <a
            href="/leaderboards"
            className="h-8 items-center justify-center text-sm font-medium px-3 py-2 inline-flex text-secondary-foreground transition-opacity hover:opacity-70 duration-200"
          >
            Leaderboard
          </a>
        </ul>
        <a
          className="text-sm font-medium pr-2 border-l-[1px] border-border pl-4 hidden md:block"
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
    <div className="flex min-h-screen flex-col container">
      <Header />
      <main className="flex-1">{children}</main>
      <footer>
        <h5 className="text-[#161616] text-[500px] leading-none text-center pointer-events-none mx-12">
          wakana
        </h5>
      </footer>
    </div>
  );
}
