"use client";

import React from "react";

export const PulseIndicator = () => {
  const [pulsing, setPulsing] = React.useState(false);

  React.useEffect(() => {
    const pulseInterval = setInterval(() => {
      setPulsing(!pulsing);
    }, 10000); // Adjust the interval as needed (milliseconds)

    return () => clearInterval(pulseInterval);
  }, []);

  React.useEffect(() => {
    const pulseInterval = setInterval(() => {
      setPulsing(false);
    }, 10000); // Adjust the interval as needed (milliseconds)

    return () => clearInterval(pulseInterval);
  }, []);

  return (
    <div
      className={`h-3 w-3 border-double outline outline-3 outline-white mr-1 rounded-full bg-red-500 ${
        pulsing ? "pulse" : ""
      }`}
    ></div>
  );
};
