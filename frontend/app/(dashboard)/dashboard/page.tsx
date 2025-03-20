import { QuestionMarkCircledIcon } from "@radix-ui/react-icons";
import { format, subDays } from "date-fns";
import { Metadata } from "next";
import Link from "next/link";

import { fetchData } from "@/actions";
import { ActivityCategoriesChart } from "@/components/charts/ActivityCategoriesChart";
import { DailyCodingSummaryOverTime } from "@/components/charts/DailyCodingSummaryOverTime";
import { WeekdaysBarChart } from "@/components/charts/WeekdaysBarChart";
import { WGaugeChart } from "@/components/charts/WGaugeChart";
import { WPieChart } from "@/components/charts/WPieChart";
import DashboardStatsSummary from "@/components/dashboard-stats-summary";
import DeveloperActivityChart from "@/components/developer-activity-chart-v2";
import { ProjectCard } from "@/components/project-card";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { SummariesApiResponse } from "@/lib/types";
import {
  convertSecondsToHoursAndMinutes,
  makePieChartDataFromRawApiResponse,
} from "@/lib/utils";

export const metadata: Metadata = {
  title: "Dashboard",
  description: "Wakana main dashboard for analytics.",
};

export default async function Dashboard({
  searchParams,
}: {
  searchParams: Record<string, any>;
}) {
  const {
    start = format(subDays(new Date(), 6), "yyyy-MM-dd"),
    end = format(new Date(), "yyyy-MM-dd"),
  } = searchParams;

  const url = `compat/wakatime/v1/users/current/summaries?${new URLSearchParams(
    { start, end }
  )}`;

  const durationData = await fetchData<SummariesApiResponse>(url, true);
  if (!durationData) {
    throw Error("Internal Server error");
  }
  const projects = durationData
    ? makePieChartDataFromRawApiResponse(durationData.data, "projects")
    : [];
  return (
    <div className="my-6">
      {durationData && (
        <main className="main-dashboard space-y-3">
          <DashboardStatsSummary
            searchParams={searchParams}
            data={durationData}
          />

          <section className="charts-grid-top">
            <div className="chart-box min-h-52">
              <DailyCodingSummaryOverTime data={durationData.data} />
            </div>
            <div className="chart-box min-h-52">
              <DeveloperActivityChart
                writePercentage={durationData.write_percentage}
                totalSeconds={+durationData.cumulative_total.seconds}
              />
            </div>
            <div className="chart-box min-h-52">
              <WGaugeChart
                data={durationData.data}
                dailyAverage={durationData.daily_average}
              />
            </div>
          </section>

          <div className="my-5 space-y-3">
            <div className="charts-grid">
              <div className="chart-box">
                <WPieChart
                  innerRadius={44.45}
                  title="Editors"
                  colorNamespace="editors"
                  data={makePieChartDataFromRawApiResponse(
                    durationData.data,
                    "editors"
                  )}
                  durationSubtitle="Editors used over the "
                />
              </div>
              <div className="chart-box">
                <WPieChart
                  title="Languages"
                  data={makePieChartDataFromRawApiResponse(
                    durationData.data,
                    "languages"
                  )}
                  colorNamespace="languages"
                  durationSubtitle="Languages used over the "
                />
              </div>
            </div>
            <div className="charts-grid">
              <div className="chart-box">
                <WPieChart
                  innerRadius={44.45}
                  title="Operating Systems"
                  data={makePieChartDataFromRawApiResponse(
                    durationData.data,
                    "operating_systems"
                  )}
                  colorNamespace="operating_systems"
                  durationSubtitle="Operating Systems used over the "
                />
              </div>
              <div className="chart-box">
                <WPieChart
                  innerRadius={44.45}
                  title="Machines"
                  data={makePieChartDataFromRawApiResponse(
                    durationData.data,
                    "machines"
                  )}
                  colorNamespace="machines"
                  durationSubtitle="Machines used over the "
                />
              </div>
            </div>
            <div className="charts-grid">
              <div className="chart-box">
                <ActivityCategoriesChart data={durationData.data} />
              </div>
              <div className="chart-box">
                <WeekdaysBarChart
                  data={durationData.data}
                  durationSubtitle="Average time per weekday over the "
                />
              </div>
            </div>
          </div>

          <div className="my-5">
            <div className="flex items-baseline gap-1 align-middle">
              <h1 className="text-2xl mb-4">Projects</h1>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <QuestionMarkCircledIcon />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Projects worked on over the last 7 days</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <div className="w-full">
              <div className="three-grid" id="projects">
                {projects.map((project: { key: string; total: number }) => (
                  <ProjectCard
                    key={project.key}
                    title={project.key}
                    duration={convertSecondsToHoursAndMinutes(
                      project.total,
                      true
                    )}
                  />
                ))}
                {projects.length === 0 && (
                  <div className="project-wrapper">
                    <a className="project-card">
                      <div className="project-content">
                        <h3>No projects available yet</h3>
                        <p>
                          Check your plugin status to see if any of your plugins
                          have been detected.
                        </p>
                        <p className="mt-2 cursor-pointer">
                          You may also checkout{" "}
                          <Link href="/installation" className="underline">
                            Installation Guide
                          </Link>
                        </p>
                      </div>
                    </a>
                  </div>
                )}
              </div>
            </div>
          </div>
        </main>
      )}
      {/* {!durationData && <h3>Error fetching dashboard stats</h3>} */}
    </div>
  );
}
