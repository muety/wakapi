import colorLib, { Color, RGBA } from "@kurkle/color";
import { type ClassValue, clsx } from "clsx";
import { format } from "date-fns";
import dayjs from "dayjs";
import duration from "dayjs/plugin/duration";
import relativeTime from "dayjs/plugin/relativeTime";
import { twMerge } from "tailwind-merge";

import { COLORS, SAMPLE_COLORS } from "../constants";
import { Category, SummariesResponse } from "../types";

dayjs.extend(duration);
dayjs.extend(relativeTime);

export const preserveNewLine = (
  text: string | undefined | null,
  fallback = ""
) => {
  if (!text) return fallback;
  return text.replace("\n", "<br />");
};

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function transparentize(
  value: string | number[] | Color | RGBA,
  opacity?: number | undefined
) {
  const alpha = opacity === undefined ? 0.5 : 1 - opacity;
  return colorLib(value).alpha(alpha).rgbString();
}

export function makePieChartDataFromRawApiResponse(
  responses: SummariesResponse[],
  key:
    | "categories"
    | "editors"
    | "operating_systems"
    | "languages"
    | "machines"
    | "projects"
) {
  const pieData: Record<string, number> = {};
  for (const response of responses) {
    for (const keyData of response[key]) {
      if (pieData[keyData.name]) {
        pieData[keyData.name] += keyData.total_seconds;
      } else {
        pieData[keyData.name] = keyData.total_seconds;
      }
    }
  }
  return Object.entries(pieData).map(([key, total]) => ({
    key,
    total,
    color: "",
  }));
}

export function deTransparentize(rgbaString: string) {
  // Use a regular expression to extract the RGBA components
  const rgbaRegex = /rgba?\((\d+),\s*(\d+),\s*(\d+),\s*(\d*\.?\d+)?\)/;
  const match = rgbaString.match(rgbaRegex);

  if (!match) {
    return { color: rgbaString };
    // throw new Error("Invalid RGBA string");
  }

  const r = parseInt(match[1]);
  const g = parseInt(match[2]);
  const b = parseInt(match[3]);
  const a = parseFloat(match[4] || "1");

  // Calculate the original opacity
  const opacity = 1 - a;

  // Create the original RGB color string
  const rgbString = `rgb(${r}, ${g}, ${b})`;

  return {
    color: rgbString,
    opacity: opacity,
  };
}

export function brightenHexColor(hex: string, percent: number) {
  // Parse the hex color to get RGB components
  let r = parseInt(hex.slice(1, 3), 16);
  let g = parseInt(hex.slice(3, 5), 16);
  let b = parseInt(hex.slice(5, 7), 16);

  // Calculate brightness adjustment
  const amount = Math.round(255 * (percent / 100));

  // Brighten the color by increasing RGB values
  r = Math.min(255, r + amount);
  g = Math.min(255, g + amount);
  b = Math.min(255, b + amount);

  // Convert back to hex format
  const brightenedHex = `#${r.toString(16).padStart(2, "0")}${g
    .toString(16)
    .padStart(2, "0")}${b.toString(16).padStart(2, "0")}`;

  return brightenedHex;
}

export function getRandomBrightColor(format = "hex") {
  // Function to generate a random integer between min and max (inclusive)
  const randomInt = (min: number, max: number) =>
    Math.floor(Math.random() * (max - min + 1)) + min;

  // Function to convert an RGB value (0-255) to a two-digit hex string
  const rgbToHex = (value: number) => value.toString(16).padStart(2, "0");

  // Generate random, slightly desaturated base RGB values (between 100 and 255)
  let r = randomInt(100, 255);
  let g = randomInt(100, 255);
  let b = randomInt(100, 255);

  // Increase saturation by adjusting towards the brightest value (255)
  const maxComponent = Math.max(r, g, b);
  r += Math.round((maxComponent - r) * 0.5); // Increase by half the difference to max
  g += Math.round((maxComponent - g) * 0.5);
  b += Math.round((maxComponent - b) * 0.5);

  // Ensure values stay within 0-255 range
  r = Math.min(Math.max(r, 0), 255);
  g = Math.min(Math.max(g, 0), 255);
  b = Math.min(Math.max(b, 0), 255);

  // Return the color in the desired format
  if (format === "hex") {
    return `#${rgbToHex(r)}${rgbToHex(g)}${rgbToHex(b)}`;
  } else if (format === "rgb") {
    return `rgb(${r}, ${g}, ${b})`;
  } else {
    throw new Error('Invalid color format. Choose "hex" or "rgb".');
  }
}

const selected = new Set();
const machineColors: Record<string, string> = {};

export function getRandomColor() {
  const color_group = SAMPLE_COLORS;
  const size = color_group.length;
  const randoms =
    new Array(100)
      .fill(100)
      .map(() => Math.floor(Math.random() * size))
      .reduce((prev, cur) => prev + cur, 0) & size;
  let color = color_group[Math.floor(randoms)];

  while (selected.has(color)) {
    color = color_group[Math.floor(Math.random() * size)];

    selected.add(color);
  }
  return color;
}

export function getRandomBrightColor2(format = "hex") {
  // Function to generate a random integer between min and max (inclusive)
  const randomInt = (min: number, max: number) =>
    Math.floor(Math.random() * (max - min + 1)) + min;

  // Function to convert an RGB value (0-255) to a two-digit hex string
  const rgbToHex = (value: number) => value.toString(16).padStart(2, "0");

  // Generate random bright RGB values (between 150 and 255)
  const r = randomInt(150, 255);
  const g = randomInt(150, 255);
  const b = randomInt(150, 255);

  // Return the color in the desired format
  if (format === "hex") {
    return `#${rgbToHex(r)}${rgbToHex(g)}${rgbToHex(b)}`;
  } else if (format === "rgb") {
    return `rgb(${r}, ${g}, ${b})`;
  } else {
    throw new Error('Invalid color format. Choose "hex" or "rgb".');
  }
}

export function convertToHoursAndMinutes(
  valueInMilliSeconds: number,
  verbose = false
) {
  const value = valueInMilliSeconds / 1000;
  const minutes = Math.floor(value % 60);
  const hours = Math.floor(value / 60);

  const hour_symbol = verbose ? " hrs" : "h";
  const minutes_symbol = verbose ? " mins" : "m";

  return `${hours}${hour_symbol} ${minutes}${minutes_symbol}`;
}

export function convertSecondsToHoursAndMinutes(
  value: number,
  verbose = false
) {
  if (value === 0 || !value) {
    return "0h 0m";
  }
  const valueInMinutes = value / 60;
  const hours = Math.floor(valueInMinutes / 60);
  const minutes = Math.floor(valueInMinutes - hours * 60);
  const seconds = Math.floor(value - (minutes + hours * 60) * 60);

  const hour_symbol = verbose ? " hrs" : "h";
  const minutes_symbol = verbose ? " mins" : "m";

  const hour_label = hours > 0 ? `${hours}${hour_symbol}` : "";
  const minutes_label = minutes > 0 ? `${minutes}${minutes_symbol}` : "";
  const seconds_label = seconds > 0 && !hours && !minutes ? `${seconds}s` : "";

  return `${hour_label} ${minutes_label} ${seconds_label}`;
}

export function getMachineColor(machineId: string) {
  if (machineColors[machineId]) {
    return machineColors[machineId];
  }
  const color = getRandomColor();
  machineColors[machineId] = color;
  return color;
}

export const convertSecondsToHours = (seconds: number) => {
  if (seconds === 0) return seconds;
  return (seconds / 60 / 60).toFixed(2) + "hrs";
};

export function getEntityColor(namespace: string, key: string) {
  if (namespace === "machine") {
    return getMachineColor(key);
  }
  const group = COLORS[namespace];
  return group ? group[key] : getRandomColor();
}

export function prepareDailyCodingData(arg: SummariesResponse) {
  const name = format(new Date(arg.range.start), "EEE LLL do");
  const date = format(new Date(arg.range.start), "yyyy-MM-dd");
  const amalgamated = arg.projects.reduce(
    (prev, curr) => ({
      [curr.name]: curr.total_seconds,
      ...prev,
      total: prev.total + curr.total_seconds,
    }),
    { name, total: 0, date }
  );
  return amalgamated;
}

export function normalizeChartData(
  rawChartData: Record<string, number>[]
): [any, string[]] {
  const projects = new Set<string>();
  rawChartData.forEach((d) =>
    Object.keys(d).forEach(
      (key) => !["name", "total"].includes(key) && projects.add(key)
    )
  );
  for (const data of rawChartData) {
    for (const projectKey of Array.from(projects)) {
      if (!data[projectKey]) {
        data[projectKey] = 0;
      }
    }
  }
  return [rawChartData, Array.from(projects)];
}

export function getUniqueProjects(
  rawChartData: {
    name: string;
    total: number;
  }[]
): string[] {
  const projects = new Set<string>();
  rawChartData.forEach((d) =>
    Object.keys(d).forEach(
      (key) => !["name", "total", "date"].includes(key) && projects.add(key)
    )
  );
  return Array.from(projects);
}

function getDaySummary(categories: Category[]) {
  return categories.reduce(
    (prev: Record<string, any>, cur: { name: any; total_seconds: any }) => ({
      ...prev,
      [cur.name]: cur.total_seconds,
    }),
    {}
  );
}

export function makeCategorySummaryData(
  rawSummaries: SummariesResponse[]
): [any[], Record<string, any>] {
  const groupedTotals: Record<string, any> = {};
  const grouped = rawSummaries.map((summary: SummariesResponse) => {
    const daySummary: Record<string, any> = getDaySummary(summary.categories);
    Object.keys(daySummary).forEach((key) => {
      if (groupedTotals[key]) {
        groupedTotals[key] += daySummary[key];
      } else {
        groupedTotals[key] = daySummary[key];
      }
    });
    return {
      ...daySummary,
      name: format(summary.range.start, "LLL do"),
    };
  });
  return [grouped, groupedTotals];
}

export function mergeDailySummary(
  first: Record<string, number>,
  second: Record<string, number>
) {
  const result: Record<string, number> = {};
  Object.keys(first).forEach((key) => {
    const otherValue = second[key] || 0;
    result[key] = otherValue + first[key];
  });
  return result;
}

export function makeCategorySummaryDataForWeekdays(
  rawSummaries: SummariesResponse[]
): [any[], Record<string, any>] {
  const groupedTotals: Record<string, any> = {};
  const summaryByDay: Record<string, any> = {};
  for (const summary of rawSummaries) {
    const daySummary: Record<string, any> = getDaySummary(summary.categories);
    Object.keys(daySummary).forEach((key) => {
      if (groupedTotals[key]) {
        // eslint-disable-next-line @typescript-eslint/no-unused-expressions
        groupedTotals[key] + daySummary[key];
      } else {
        groupedTotals[key] = daySummary[key];
      }
    });
    const day = format(summary.range.start, "EEEE");
    if (summaryByDay[day]) {
      summaryByDay[day] = mergeDailySummary(daySummary, summaryByDay[day]);
    } else {
      summaryByDay[day] = daySummary;
    }
  }

  const grouped = Object.entries(summaryByDay).map(([key, value]) => ({
    name: key,
    ...value,
  }));
  return [grouped, groupedTotals];
}

export function humanizeDate(dateString: string) {
  const targetDate = new Date(dateString);
  const now = dayjs();
  const target = dayjs(targetDate);
  const diffDuration = dayjs.duration(target.diff(now));
  return diffDuration.humanize(true);
}

export const formatNumber = (
  value: number,
  options?: Intl.NumberFormatOptions
) => {
  if (!options?.currency) {
    delete options?.currency;
  }
  if (
    typeof Intl === "object" &&
    Intl &&
    typeof Intl.NumberFormat === "function"
  ) {
    return new Intl.NumberFormat("en-US", {
      ...options,
    }).format(value);
  }

  return value.toString();
};

export function formatCurrency(value: number, currency: string) {
  return formatNumber(value, {
    style: "currency",
    currency,
  });
}

export const getHours = (seconds: number) => {
  return seconds / 3600;
};
