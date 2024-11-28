interface AuthLayoutProps {
  children: React.ReactNode;
}

import { HeroBrand } from "@/components/hero-brand";
import { cn } from "@/lib/utils";

export default async function Page({ children }: AuthLayoutProps) {
  return (
    <div className="container flex h-screen min-h-screen w-screen flex-col items-center justify-center">
      <div className={cn("md:absolute flex items-start md:left-8 md:top-8")}>
        <HeroBrand
          imgHeight={30}
          imgWidth={30}
          fontSize="42px"
          lineHeight="35px"
        />
      </div>
      <div className="mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]">
        <div className="flex flex-col justify-center justify-items-center space-y-2 text-center align-middle">
          {children}
        </div>
      </div>
    </div>
  );
}
