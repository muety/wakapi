import { SparklesIcon } from "lucide-react";
import Image from "next/image";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
// import Blur1 from "@/public//dashboard.webp";
import { cn } from "@/lib/utils";
import { FadeOnView } from "@/components/fade-on-view";
import Link from "next/link";

export function Hero() {
  return (
    <section className="px-6 flex flex-col justify-center text-center relative">
      <div className="min-h-[30vh] py-8">
        <FadeOnView>
          <Badge variant="secondary" className="w-fit mx-auto">
            <SparklesIcon className="w-4 h-4 mr-2" />
            Beta Testing
          </Badge>
        </FadeOnView>
        <FadeOnView delay={0.2}>
          <h1 className="heading vertical-gradient lg:text-7xl sm:flex-center gap-2 sm:gap-4 py-4 md:py-6">
            Observe your work in real time
          </h1>
        </FadeOnView>
        <FadeOnView delay={0.4}>
          <p className="text-muted-foreground md:text-lg">
            Developer dashboards for insights into your work habits
          </p>
        </FadeOnView>
        <div className="flex-gap justify-center">
          <FadeOnView delay={0.6} className="space-x-2 mt-12">
            {/* <Button className="rounded-full shadow-lg">Get Started</Button> */}
            <Link href="/auth/signin" className={cn("font-heading hero-cta")}>
              Try it for free
            </Link>
          </FadeOnView>
        </div>
      </div>
      <FadeOnView
        delay={1}
        className="hero-border-animation max-w-screen-xl mx-auto mt-16 p-[1px] rounded-[1rem] bg-ring"
        style={{
          maskImage: "linear-gradient(to bottom, black 30%, transparent 90%)",
        }}
      >
        <div
          className={cn(
            "rounded-[1rem] overflow-hidden p-2 z-10",
            "bg-background"
          )}
        >
          <div className="z-10">
            <Image
              src="/dashboard.webp"
              alt="App image"
              width={1920}
              height={1080}
              className="rounded-[12px] overflow-hidden z-10 border relative"
            />
            <Image
              src="/bg-blur-1.webp"
              alt="background blur"
              width={1920}
              height={1080}
              className="opacity-30 absolute"
            />
          </div>
        </div>
      </FadeOnView>
    </section>
  );
}
