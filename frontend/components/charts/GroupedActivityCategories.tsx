"use client";

import { Bar, BarChart, ResponsiveContainer, XAxis } from "recharts";

const data = [
  {
    name: "Apr 2nd",
    total: Math.floor(Math.random() * 5000) + 1000,
  },
  {
    name: "Apr 3rd",
    total: Math.floor(Math.random() * 5000) + 1000,
  },
  {
    name: "Apr 4th",
    total: Math.floor(Math.random() * 5000) + 1000,
  },
  {
    name: "Apr 5th",
    total: Math.floor(Math.random() * 5000) + 1000,
  },
  {
    name: "Apr 6th",
    total: Math.floor(Math.random() * 5000) + 1000,
  },
  {
    name: "Apr 7th",
    total: Math.floor(Math.random() * 5000) + 1000,
  },
  {
    name: "Apr 8th",
    total: Math.floor(Math.random() * 5000) + 1000,
  },
];

export function BarExample() {
  return (
    <ResponsiveContainer width="100%" height={350}>
      <BarChart data={data} title="Weekdays">
        <XAxis
          dataKey="name"
          stroke="#888888"
          fontSize={12}
          tickLine={false}
          axisLine={false}
        />
        {/* <YAxis
          stroke="#888888"
          fontSize={12}
          tickLine={false}
          axisLine={false}
          tickFormatter={(value) => `$${value}`}
        /> */}
        <Bar dataKey="total" fill="#315340" radius={[4, 4, 0, 0]} />
        <Bar dataKey="total" fill="green" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  );
}
