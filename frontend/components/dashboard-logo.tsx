"use client";

import Image from "next/image";
import { cn } from "@/lib/utils";

import React from "react";
import { Rubik } from "next/font/google";
import CustomLink from "./custom-link";
const rubik = Rubik({ subsets: ["latin"] });

export interface HeroBrandProps {
  imgWidth?: number;
  imgHeight?: number;
  fontSize?: string;
  lineHeight?: string;
  white?: boolean;
}

export function DashboardLogo() {
  // TODO: Rendering hydration issue to do with nesting of a block level component, Image, inside
  //an anchor tag
  return (
    <div className="flex gap-4 items-center justify-center md:justify-start">
      <CustomLink
        href="/"
        className={cn("flex --font-heading items-center", rubik.className)}
      >
        <p className="text-xl font-bold">Wakana</p>
      </CustomLink>
    </div>
  );
}
