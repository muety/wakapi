"use client";

import { useEffect, useState } from "react";

export function Background() {
  return (
    <div
      className="-mt-12"
      style={{
        backgroundImage: `url(/bg1.svg)`,
        backgroundSize: "cover",
        backgroundPosition: "center",
        width: "100%",
        height: "100%",
        position: "absolute",
        top: 0,
        left: 0,
      }}
    ></div>
  );
}
