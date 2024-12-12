"use client";

import { startOfDay } from "date-fns";
import React from "react";

import { fetchData } from "@/actions";
import { useClientSession } from "@/lib/session";
import { SummariesApiResponse } from "@/lib/types";

import { PulseIndicator } from "./pulse-indicator";

export function CurrentWorkTime() {
  const { data: session, isLoading } = useClientSession();

  const [todaysCodingTime, setTodaysCodingTime] =
    React.useState("- hrs - mins");

  const token = React.useMemo(() => session?.token || "", [session]);

  const fetchSummary = async () => {
    const start = startOfDay(new Date()).toISOString();
    const end = new Date().toISOString();
    const url = `compat/wakatime/v1/users/current/summaries?${new URLSearchParams(
      { start, end }
    )}`;
    const durationData = await fetchData<SummariesApiResponse>(url);
    if (durationData)
      setTodaysCodingTime(durationData.cumulative_total.text || "");
  };

  React.useEffect(() => {
    if (token) fetchSummary();
  }, [session, isLoading, token]);

  React.useEffect(() => {
    const pulseInterval = setInterval(() => {
      fetchSummary();
    }, 20000);

    return () => clearInterval(pulseInterval);
  }, [session]);

  return (
    <div
      className="flex items-center justify-center rounded-lg border border-border px-3 align-middle text-slate-100 shadow"
      style={{
        paddingLeft: "10px",
        paddingRight: "10px",
        fontWeight: "bold",
        lineHeight: "2rem",
        fontFamily: "monaco",
        fontSize: "0.95rem",
      }}
    >
      <div className="flex items-center justify-center gap-1">
        <PulseIndicator />
        {todaysCodingTime}
      </div>
    </div>
  );
}
