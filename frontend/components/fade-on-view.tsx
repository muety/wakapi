"use client";

import * as React from "react";
import { useEffect, useState } from "react";
import { useInView } from "react-intersection-observer";

import { cn } from "@/lib/utils";

interface FadeOnViewProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
  delay?: number;
  offset?: number;
  className?: string;
}

export function FadeOnView({
  children,
  delay,
  offset,
  className,
  ...props
}: FadeOnViewProps) {
  const [viewed, setViewed] = useState(false);

  const { ref, inView } = useInView({
    rootMargin: `0% 0% ${-15 + (offset ?? 0)}% 0%`,
    initialInView: false,
  });

  useEffect(() => {
    if (inView) setViewed(true);
  }, [inView]);

  return (
    <div
      ref={ref}
      data-inview={viewed}
      className={cn(
        // "bg-red-500",
        "transition-all",
        "duration-700",
        "data-[inview=false]:opacity-0",
        "data-[inview=false]:translate-y-10",
        "data-[inview=true]:opacity-100",
        "data-[inview=true]:translate-y-0",
        className
      )}
      style={{
        transitionDelay: `${delay ?? 0}s`,
        ...props.style,
      }}
    >
      {children}
    </div>
  );
}
