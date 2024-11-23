"use client";

import React from "react";
import { startOfDay } from "date-fns";
import { fetchData } from "@/actions";
import { useClientSession } from "@/lib/session";
import { PulseIndicator } from "./pulse-indicator";
import { SummariesApiResponse } from "@/lib/types";

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
    durationData &&
      setTodaysCodingTime(durationData.cumulative_total.text || "");
  };

  React.useEffect(() => {
    token && fetchSummary();
  }, [session, isLoading]);

  React.useEffect(() => {
    const pulseInterval = setInterval(() => {
      fetchSummary();
    }, 20000);

    return () => clearInterval(pulseInterval);
  }, [session]);

  return (
    <div
      className="flex items-center align-middle justify-center border px-3 border-border rounded-lg shadow text-slate-100"
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
