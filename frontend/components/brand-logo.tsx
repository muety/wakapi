"use client";

import Image from "next/image";

export function BrandLogo({
  height = 50,
  width = 50,
  alt = "logo",
  white = false,
}: {
  height?: number;
  width?: number;
  alt?: string;
  white?: boolean;
}) {
  const src = white ? "/white-logo.svg" : "/logo.svg";
  return <Image src={src} height={height} width={width} alt="logo" />;
}
