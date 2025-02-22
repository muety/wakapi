import TimeTrackingVisualization from "@/components/day-dashboard/day";
import { SAMPLE_DURATIONS3 } from "@/lib/constants/durations";

export default function DayPage() {
  return <TimeTrackingVisualization width={1224} data={SAMPLE_DURATIONS3} />;
}
