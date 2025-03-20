"use client";

import { getSelectedPeriodLabel } from "@/lib/utils";
import { Code, GitBranch, Clock } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { DashboardPeriodSelectorV2 } from "./dashboard-period-selector";

export default function DashboardStatsSummary({
  codingTime,
  searchParams,
}: {
  codingTime: string;
  searchParams: Record<string, any>;
}) {
  const selectedPeriodLabel = getSelectedPeriodLabel(searchParams || {});

  return (
    <div className="w-full text-white">
      <div className="max-w-7xl mx-auto">
        <div className="flex flex-col space-y-6">
          {/* Header with time range selector */}
          <div className="flex items-center justify-between border p-4 rounded-sm">
            <div>
              <h2 className="text-gray-400 text-sm font-medium mb-1">
                Total Coding Time
              </h2>
              <div className="flex items-baseline">
                <span className="text-3xl font-bold bg-gradient-to-r from-blue-500 to-purple-500 bg-clip-text text-transparent">
                  {codingTime}
                </span>
                <span className="text-gray-400 ml-2 text-sm">
                  over {selectedPeriodLabel.toLowerCase()}
                </span>
              </div>
            </div>

            <DashboardPeriodSelectorV2 searchParams={searchParams} />
          </div>
        </div>
      </div>
    </div>
  );
}

export function StatsSection() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
      <Card className="border-gray-800 overflow-hidden">
        <CardContent className="p-0">
          <div className="p-4">
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-gray-400 text-sm font-medium">Projects</h3>
              <Code className="h-4 w-4 text-blue-400" />
            </div>
            <p className="text-2xl font-bold">12</p>
          </div>
          <div className="h-1 w-full bg-gradient-to-r from-blue-500 to-blue-600"></div>
        </CardContent>
      </Card>

      <Card className="border-gray-800 overflow-hidden">
        <CardContent className="p-0">
          <div className="p-4">
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-gray-400 text-sm font-medium">Languages</h3>
              <Code className="h-4 w-4 text-purple-400" />
            </div>
            <p className="text-2xl font-bold">8</p>
          </div>
          <div className="h-1 w-full bg-gradient-to-r from-purple-500 to-purple-600"></div>
        </CardContent>
      </Card>

      <Card className="border-gray-800 overflow-hidden">
        <CardContent className="p-0">
          <div className="p-4">
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-gray-400 text-sm font-medium">Commits</h3>
              <GitBranch className="h-4 w-4 text-teal-400" />
            </div>
            <p className="text-2xl font-bold">47</p>
          </div>
          <div className="h-1 w-full bg-gradient-to-r from-teal-500 to-teal-600"></div>
        </CardContent>
      </Card>

      <Card className="border-gray-800 overflow-hidden">
        <CardContent className="p-0">
          <div className="p-4">
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-gray-400 text-sm font-medium">Avg. Daily</h3>
              <Clock className="h-4 w-4 text-amber-400" />
            </div>
            <p className="text-2xl font-bold">5.4h</p>
          </div>
          <div className="h-1 w-full bg-gradient-to-r from-amber-500 to-amber-600"></div>
        </CardContent>
      </Card>
    </div>
  );
}
