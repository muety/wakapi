"use client";

import React from "react";

export const PulseIndicator = () => {
  const [pulsing, setPulsing] = React.useState(false);

  React.useEffect(() => {
    const pulseInterval = setInterval(() => {
      setPulsing(!pulsing);
    }, 10000);

    return () => clearInterval(pulseInterval);
  });

  React.useEffect(() => {
    const pulseInterval = setInterval(() => {
      setPulsing(false);
    }, 10000); // Adjust the interval as needed (milliseconds)

    return () => clearInterval(pulseInterval);
  }, [pulsing]);

  return (
    <div
      className={`outline-3 mr-1 size-3 rounded-full border-double bg-red-500 outline outline-white ${
        pulsing ? "pulse" : ""
      }`}
    ></div>
  );
};
