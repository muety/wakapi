import {
  differenceInCalendarDays,
  endOfMonth,
  endOfWeek,
  format,
  parse,
  startOfMonth,
  startOfWeek,
  subDays,
  subMonths,
  subWeeks,
} from "date-fns";

export enum DashboardRangeQuery {
  Last7Days = "Last 7 Days",
  Last7DaysFromYesterday = "Last 7 Days From Yesterday",
  Last14Days = "Last 14 Days",
  Last30Days = "Last 30 Days",
  ThisWeek = "This Week",
  LastWeek = "Last Week",
  LastMonth = "Last Month",
}

function formatDate(date: Date): string {
  return format(date, "yyyy-MM-dd");
}

function formatDashboardDuration(start: Date, end: Date) {
  return `?start=${formatDate(start)}&end=${formatDate(end)}`;
}

export function buildQueryForRangeQuery(query: DashboardRangeQuery): string {
  let yesterday: Date;
  let lastMonth: Date;
  let lastWeek: Date;

  switch (query) {
    case DashboardRangeQuery.Last7Days:
      return formatDashboardDuration(subDays(new Date(), 7), new Date());
    case DashboardRangeQuery.Last7DaysFromYesterday:
      yesterday = subDays(new Date(), 1);
      return formatDashboardDuration(subDays(yesterday, 7), yesterday);
    case DashboardRangeQuery.Last14Days:
      return formatDashboardDuration(subDays(new Date(), 14), new Date());
    case DashboardRangeQuery.Last30Days:
      return formatDashboardDuration(subDays(new Date(), 30), new Date());
    case DashboardRangeQuery.LastMonth:
      lastMonth = subMonths(new Date(), 1);
      return formatDashboardDuration(
        startOfMonth(lastMonth),
        endOfMonth(lastMonth)
      );
    case DashboardRangeQuery.LastWeek:
      lastWeek = subWeeks(startOfWeek(new Date()), 1);
      return formatDashboardDuration(
        startOfWeek(lastWeek),
        endOfWeek(lastWeek)
      );
    case DashboardRangeQuery.ThisWeek:
      return formatDashboardDuration(startOfWeek(new Date()), new Date());
  }
}
export function parseDate(dateString: string): Date {
  const formatString = "yyyy-MM-dd";
  const parsedDate = parse(dateString, formatString, new Date());
  return parsedDate;
}

export function getSelectedPeriodLabel(searchParams: Record<string, any>) {
  const { start, end } = searchParams;
  if (!start || !end) {
    return DashboardRangeQuery.Last7Days;
  }
  const startDate = new Date(start);
  const endDate = new Date(end);
  const difference = differenceInCalendarDays(endDate, startDate);

  if (difference === 7) {
    if (endDate.getDay() === new Date().getDay()) {
      return DashboardRangeQuery.Last7Days;
    } else {
      return DashboardRangeQuery.Last7DaysFromYesterday;
    }
  }

  if (difference === 14) {
    return DashboardRangeQuery.Last14Days;
  }

  if (endDate.getDay() === endOfMonth(subMonths(new Date(), 1)).getDay()) {
    return DashboardRangeQuery.LastMonth;
  }

  if (difference === 30) {
    return DashboardRangeQuery.Last30Days;
  }

  // last week
  const endOfLastWeek = endOfWeek(subWeeks(new Date(), 1));
  if (endDate.getDay() === endOfLastWeek.getDay()) {
    return DashboardRangeQuery.LastWeek;
  }

  // this week
  if (startDate.getDay() === startOfWeek(new Date()).getDay()) {
    return DashboardRangeQuery.ThisWeek;
  }
  // last month
  return DashboardRangeQuery.Last7Days;
}
