import { cn } from "@/lib/utils";
import Link from "next/link";
import { Background } from "./background2";
import styles from "./hero.module.css";

export function Hero() {
  return (
    <div className={styles.heroContainer}>
      <section className="relative z-10 space-y-6 mb-36 pt-6 md:pt-10 top">
        <div
          className={cn(
            "container flex max-w-[64rem] flex-col items-center gap-4 text-center mb-5",
            styles.heroContainer
          )}
        >
          <h1 className="font-heading font-medium text-3xl sm:text-4xl md:text-4xl lg:text-5xl hero-title text-white mb-4">
            Observe your work in real time
            <span className="sr-only">one heartbeat at a time!</span>
          </h1>
          <h2 className="hero-subtitle">
            Developer dashboards for insights into your work habits
          </h2>
          <div className="space-x-4 rounded-sm">
            <Link href="/auth/signin" className={cn("font-heading hero-cta")}>
              Try it for free
            </Link>
          </div>
        </div>
      </section>
      <Background />
    </div>
  );
}
