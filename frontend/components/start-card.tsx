import type { LucideIcon } from "lucide-react";

interface StatCardProps {
  icon: LucideIcon;
  value: string;
  label: string;
  iconColor: string;
}

export function StatCard({
  icon: Icon,
  value,
  label,
  iconColor,
}: StatCardProps) {
  return (
    <div className="border border-[#4f5b69] rounded-xl p-6 text-center">
      <Icon className={`h-8 w-8 mx-auto mb-4 ${iconColor}`} />
      <h4 className="text-2xl font-bold mb-1">{value}</h4>
      <p className="text-gray-400">{label}</p>
    </div>
  );
}
