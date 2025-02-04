import { HamburgerMenuIcon } from "@radix-ui/react-icons";
import Image from "next/image";
import Link from "next/link";

import { CurrentWorkTime } from "./current-work-time";
import { MobileNav } from "./mobile-nav";
import { Button } from "./ui/button";
import { Drawer, DrawerContent, DrawerTrigger } from "./ui/drawer";

export function MobileHeader() {
  return (
    <header className="sticky top-0 z-30 flex min-h-16 items-center justify-between gap-4 border-b bg-background px-4 sm:static sm:h-auto sm:border-0 sm:bg-transparent sm:px-6 md:hidden">
      <Link href="/">
        <Image
          src={"/white-logo.png"}
          alt="Logo"
          width={120}
          height={65}
          className="logo-icon-white"
        />
      </Link>
      <div className="flex items-center gap-4">
        <CurrentWorkTime />
        <Drawer>
          <DrawerTrigger asChild>
            <Button size="icon" variant="outline" className="md:hidden">
              <HamburgerMenuIcon className="size-5" />
              <span className="sr-only">Toggle Menu</span>
            </Button>
          </DrawerTrigger>
          <DrawerContent className="hero-bg h-screen">
            <MobileNav />
          </DrawerContent>
        </Drawer>
      </div>
    </header>
  );
}
