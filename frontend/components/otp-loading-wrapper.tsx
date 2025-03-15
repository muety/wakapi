"use client";

import type { ReactNode } from "react";
import { Loader2 } from "lucide-react";

// Using descriptive names for loader types
type LoaderType = "spinner" | "dots" | "pulse";

interface OTPLoadingWrapperProps {
  children: ReactNode;
  loading: boolean;
  loaderType?: LoaderType;
  className?: string;
}

export function OTPLoadingWrapper({
  children,
  loading,
  loaderType = "spinner",
  className = "",
}: OTPLoadingWrapperProps) {
  return (
    <div className={`relative w-fit ${className}`}>
      {children}

      {/* Standard Spinner - Adjusted opacity */}
      {loading && loaderType === "spinner" && (
        <div className="absolute inset-0 flex items-center justify-center bg-background/60 backdrop-blur-[2px] rounded-md z-10">
          <div className="bg-background/80 p-3 rounded-full">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        </div>
      )}

      {/* Bouncing Dots - Adjusted opacity */}
      {loading && loaderType === "dots" && (
        <div className="absolute inset-0 flex items-center justify-center bg-background/60 backdrop-blur-[2px] rounded-md z-10">
          <div className="bg-background/80 px-4 py-2 rounded-full">
            <div className="flex space-x-2">
              <div
                className="w-3 h-3 bg-primary rounded-full animate-bounce"
                style={{ animationDelay: "0ms" }}
              ></div>
              <div
                className="w-3 h-3 bg-primary rounded-full animate-bounce"
                style={{ animationDelay: "150ms" }}
              ></div>
              <div
                className="w-3 h-3 bg-primary rounded-full animate-bounce"
                style={{ animationDelay: "300ms" }}
              ></div>
            </div>
          </div>
        </div>
      )}

      {/* Pulsing Glow - Adjusted opacity */}
      {loading && loaderType === "pulse" && (
        <div className="absolute inset-0 flex items-center justify-center bg-background/60 backdrop-blur-[2px] rounded-md overflow-hidden z-10">
          <div className="absolute inset-0 bg-primary/10 animate-pulse"></div>
          <div className="z-10 flex items-center justify-center space-x-2 bg-background/80 px-4 py-2 rounded-full">
            <div className="h-1 w-1 bg-primary rounded-full animate-ping"></div>
            <div
              className="h-1.5 w-1.5 bg-primary rounded-full animate-ping"
              style={{ animationDelay: "300ms" }}
            ></div>
            <div
              className="h-2 w-2 bg-primary rounded-full animate-ping"
              style={{ animationDelay: "600ms" }}
            ></div>
          </div>
        </div>
      )}
    </div>
  );
}
