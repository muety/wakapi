import { fetchData } from "@/actions/data-fetching";
import TimeTrackingVisualization from "@/components/day-dashboard/day";
import { DurationData } from "@/lib/types";

interface iProps {
  params: { date: string };
}

export default async function DayPage({ params: { date } }: iProps) {
  console.log("date");
  const url = `/v1/users/current/durations?${new URLSearchParams({
    date,
  })}`;
  const durationData = await fetchData<DurationData>(url);
  if (!durationData) {
    return (
      <h3 className="text-center text-red-500">
        Error fetching dashboard stats
      </h3>
    );
  }
  return <TimeTrackingVisualization width={1224} data={durationData} />;
}
