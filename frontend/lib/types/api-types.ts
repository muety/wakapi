import { Client } from "@/components/clients-table";

export interface Duration {
  time: number;
  project: string;
  duration: number;
  color: string | null;
}

export interface Time {
  digital: string;
  hours: number;
  minutes: number;
  seconds: number;
  text: string;
  total_seconds: number;
}

export interface InvoiceLineItem {
  title: string;
  total_seconds: number;
  auto_generated: boolean;
}

export interface PreparedInvoiceData {
  preamble: string;
  main_message: string;
  from: string;
  to: string;
  tax: number;
  client: Client;
  date: Date;
  line_items: InvoiceLineItem[];
}

export interface Category {
  name: string;
  percent: number;
  digital: string;
  Time: Time; // Use the Time interface here
  total_seconds: number;
  text: string;
}

export interface Project {
  id?: string;
  name: string;
  percent: number;
  digital: string;
  Time: Time; // Use the Time interface here
  total_seconds: number;
  text: string;
}

export interface Range {
  date: string;
  end: string;
  start: string;
  text: string;
  timezone: string;
}

export interface Entity {
  digital: string;
  hours: string;
  minutes: string;
  name: string;
  percent: number;
  seconds: string;
  text: string;
  total_seconds: number;
}

export interface SummariesResponse {
  categories: Category[];
  dependencies: string[];
  editors: Category[];
  languages: Category[];
  machines: Category[];
  operating_systems: Category[];
  projects: Project[];
  grand_total: Time;
  range: Range;
  entities: Entity[];
  branches: Entity[];
}

export interface CumulativeTotal {
  decimal: string;
  digital: string;
  seconds: string;
  text: string;
}

export interface DailyAverage {
  days_including_holidays: number;
  days_minus_holidays: number;
  holidays: number;
  seconds: number;
  seconds_including_other_language: number;
  text: string;
  text_including_other_language: string;
}

export interface SummariesApiResponse {
  data: SummariesResponse[];
  start: string;
  end: string;
  cumulative_total: CumulativeTotal;
  daily_average: DailyAverage;
}

export interface GoalData {
  id: string;
  user_id: string;
  created_at: string; // Can be improved to Date type
  updated_at: string; // Can be improved to Date type
  snooze_until: number;
  target_direction: "more" | "less";
  seconds: number;
  improve_by_percent: number;
  delta: "day" | "week" | "month"; // Allows for future delta unit expansion
  type: string;
  title: string;
  custom_title: string | null;
  cumulative_status: string;
  status: string;
  is_snoozed: boolean;
  is_enabled: boolean;
  languages: null;
  projects: null;
  editors: null;
  categories: null;
  chart_data: GoalChartData[];
}

export interface GoalChartData {
  actual_seconds: number;
  actual_seconds_text: string;
  goal_seconds: number;
  goal_seconds_text: string;
  range_status: "pending" | string; // Allows for future range status expansion
  range_status_reason: string;
  range_status_reason_short: string;
  range: Range;
  range_text?: string; // hack-alert
}

export interface Invoice {
  id: string;
  name: string;
  amount: number;
  origin: string;
  destination: string;
  heading: string;
  final_message: string;
  invoice_summary: string;
  client: Client;
  created_at: string;
  start_date: string;
  end_date: string;
  tax: number;
  line_items: InvoiceLineItem[];
}

// LEADERBOARD

export interface LanguageStats {
  name: string;
  total_seconds: number;
}

export interface RunningTotal {
  total_seconds: number;
  human_readable_total: string;
  daily_average: number;
  human_readable_daily_average: string;
  languages: LanguageStats[];
}

export interface User {
  id: string;
  display_name: string;
  full_name: string;
  email: string;
  is_email_public: boolean;
  is_email_confirmed: boolean;
  timezone: string;
  last_heartbeat_at: string;
  last_project: string;
  last_plugin_name: string;
  username: string;
  website: string;
  created_at: string;
  modified_at: string;
  photo: string;
}

export type UserProfile = {
  id: string;
  email: string;
  location: string;
  created_at: string; // ISO date string
  last_logged_in_at: string; // ISO date string
  email_verified: boolean;
  public_leaderboard: boolean;
  hireable: boolean;
  show_email_in_public: boolean;
  heartbeats_timeout_sec: number;
  name: string;
  username: string;
  bio: string;
  github_handle: string;
  twitter_handle: string;
  linked_in_handle: string;
};

export interface DataItem {
  rank: number;
  running_total: RunningTotal;
  user: User;
}

export interface LeaderboardRange {
  end_text: string;
  end_date: string;
  start_text: string;
  start_date: string;
  name: string;
  text: string;
}

export interface LeaderboardApiResponse {
  current_user: User | null;
  data: DataItem[];
  page: number;
  total_pages: number;
  language: string;
  range: LeaderboardRange;
}
