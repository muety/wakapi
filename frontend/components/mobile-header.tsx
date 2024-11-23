import { PanelLeft } from "lucide-react";
import { MobileNav } from "./mobile-nav";
import { Button } from "./ui/button";
import { Drawer, DrawerContent, DrawerTrigger } from "./ui/drawer";
import { HamburgerMenuIcon } from "@radix-ui/react-icons";
import { HeroBrand } from "./hero-brand";

export function MobileHeader() {
  return (
    <header className="md:hidden sticky top-0 z-30 flex min-h-16 items-center justify-between gap-4 border-b bg-background px-4 sm:static sm:h-auto sm:border-0 sm:bg-transparent sm:px-6">
      <div>
        <HeroBrand
          imgHeight={25}
          imgWidth={23}
          fontSize="28px"
          lineHeight="25px"
        />
      </div>
      <Drawer>
        <DrawerTrigger asChild>
          <Button size="icon" variant="outline" className="md:hidden">
            <HamburgerMenuIcon className="h-5 w-5" />
            <span className="sr-only">Toggle Menu</span>
          </Button>
        </DrawerTrigger>
        <DrawerContent className="hero-bg h-screen">
          <MobileNav />
        </DrawerContent>
      </Drawer>
    </header>
  );
}
