import { BarChart } from "lucide-react";

interface EmptyStateProps {
  title?: string;
  message?: string;
}

const EmptyState = ({
  title = "No data to show",
  message = "May take 24 hours for data to show",
}: EmptyStateProps) => (
  <div
    className="
      flex h-full
      w-full flex-col items-center
      justify-center space-y-2 border
      border-dashed border-control py-4 text-center
    "
  >
    <BarChart />
    <div>
      <p className="text-xs text-foreground-light">{title}</p>
      <p className="text-xs text-foreground-light">{message}</p>
    </div>
  </div>
);

export default EmptyState;
