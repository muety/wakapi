import { SparklesIcon } from "lucide-react";
import Image from "next/image";
import Link from "next/link";

import { FadeOnView } from "@/components/fade-on-view";
import { Badge } from "@/components/ui/badge";
// import Blur1 from "@/public//dashboard.webp";
import { cn } from "@/lib/utils";

export function Hero() {
  return (
    <section className="relative flex flex-col justify-center px-6 text-center">
      <div className="min-h-[30vh] py-8">
        <FadeOnView>
          <Badge variant="secondary" className="mx-auto w-fit">
            <SparklesIcon className="mr-2 size-4" />
            Beta Testing
          </Badge>
        </FadeOnView>
        <FadeOnView delay={0.2}>
          <h1 className="heading vertical-gradient sm:flex-center gap-2 py-4 sm:gap-4 md:py-6 lg:text-7xl">
            Observe your work in real time
          </h1>
        </FadeOnView>
        <FadeOnView delay={0.4}>
          <p className="text-muted-foreground md:text-lg">
            Developer dashboards for insights into your work habits
          </p>
        </FadeOnView>
        <div className="flex-gap justify-center">
          <FadeOnView delay={0.6} className="mt-12 space-x-2">
            {/* <Button className="rounded-full shadow-lg">Get Started</Button> */}
            <Link href="/login" className={cn("font-heading hero-cta")}>
              Try it for free
            </Link>
          </FadeOnView>
        </div>
      </div>
      <FadeOnView
        delay={1}
        className="hero-border-animation mx-auto mt-16 max-w-screen-xl rounded-2xl bg-ring p-px"
        style={{
          maskImage: "linear-gradient(to bottom, black 30%, transparent 90%)",
          borderWidth: 0,
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
              src="/neo-dashboard-alt.png"
              alt="App image"
              width={1920}
              height={1080}
              className="relative z-10 overflow-hidden rounded-[12px] border"
            />
            <Image
              src="/bg-blur-1.webp"
              alt="background blur"
              width={1920}
              height={1080}
              className="absolute opacity-30"
            />
          </div>
        </div>
      </FadeOnView>
    </section>
  );
}
