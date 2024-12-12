"use client";

import { Rubik } from "next/font/google";
import Image from "next/image";

import { cn } from "@/lib/utils";

import CustomLink from "./custom-link";
const rubik = Rubik({ subsets: ["latin"] });

export interface HeroBrandProps {
  imgWidth?: number;
  imgHeight?: number;
  fontSize?: string;
  lineHeight?: string;
  logoType?: "white" | "black" | "none";
  inDashboard?: boolean;
}

export function HeroBrand({
  imgWidth = 36,
  imgHeight = 36,
  fontSize = "50px",
  lineHeight = "20px",
  logoType = "none",
}: HeroBrandProps) {
  const sidebar = { open: true };
  return (
    <div className="flex items-center justify-center gap-4 md:justify-start">
      <CustomLink
        href="/"
        className={cn("flex --font-heading items-center", rubik.className)}
      >
        {logoType === "none" && (
          <>
            <Image
              src={"/white-logo.png"}
              alt="Logo"
              width={imgWidth}
              height={imgHeight}
              className="logo-icon-white"
            />
            <Image
              src={"/logo.svg"}
              alt="Logo"
              width={imgWidth}
              height={imgHeight}
              className="logo-icon-black"
            />
          </>
        )}

        {logoType === "white" && (
          <Image
            src={"/white-logo.png"}
            alt="Logo"
            width={imgWidth}
            height={imgHeight}
          />
        )}

        {logoType === "black" && (
          <Image
            src={"/logo.svg"}
            alt="Logo"
            width={imgWidth}
            height={imgHeight}
            className="logo-icon-white"
          />
        )}
        {sidebar?.open && (
          <span
            className={cn("ml-1", logoType === "white" ? "text-white" : "text")}
            style={{
              fontSize,
              lineHeight,
            }}
          >
            Wakana
          </span>
        )}
      </CustomLink>
    </div>
  );
}
