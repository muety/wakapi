import React from "react";

import { cn } from "@/lib/utils";

interface VerticalSeparatorProps extends React.HTMLAttributes<HTMLDivElement> {
  /**
   * The height of the separator - defaults to "h-full" if not specified
   */
  height?: string;
  /**
   * Spacing to apply on the left and right of the separator
   */
  spacing?: string;
}

const VerticalSeparator = React.forwardRef<
  HTMLDivElement,
  VerticalSeparatorProps
>(({ className, height = "h-full", spacing = "mx-2", ...props }, ref) => {
  return (
    <div
      ref={ref}
      className={cn(
        "w-px bg-border bg-white border shrink-0",
        height,
        spacing,
        className
      )}
      {...props}
    />
  );
});

VerticalSeparator.displayName = "VerticalSeparator";

export { VerticalSeparator };
