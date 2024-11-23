"use client";

import Image from "next/image";
import { useEffect, useState } from "react";

export function Background() {
  const images = [
    "/bg1.svg",
    "/bg2.svg",
    "/bg3.svg",
    "/bg4.svg",
    "/bg5.svg",
    "/bg6.svg",
    "/bg7.svg",
  ];
  const [currentImage, setCurrentImage] = useState("bg1.svg");

  useEffect(() => {
    setInterval(() => {
      const newImage = images[Math.floor(Math.random() * images.length)];
      setCurrentImage(newImage);
    }, 10000);
  }, []);

  return (
    <div className="below">
      <Image
        src={currentImage}
        alt="Hero stats"
        width={1920}
        height={549}
        className="text-center"
        style={{ width: "1920px" }}
      />
    </div>
  );
}
