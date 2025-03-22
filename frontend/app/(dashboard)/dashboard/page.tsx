import { format, subDays } from "date-fns";
import { Metadata } from "next";
import Link from "next/link";

import { fetchData } from "@/actions";
import { ActivityCategoriesChart } from "@/components/charts/ActivityCategoriesChart";
import { DailyCodingSummaryOverTime } from "@/components/charts/DailyCodingSummaryOverTime";
import { WBarChart } from "@/components/charts/WBarChart";
import { WeekdaysBarChart } from "@/components/charts/WeekdaysBarChart";
import { WGaugeChart } from "@/components/charts/WGaugeChart";
import DashboardStatsSummary from "@/components/dashboard-stats-summary";
import DeveloperActivityChart from "@/components/developer-activity-chart-v2";
import { ProjectCard } from "@/components/project-card";
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
  const start =
    searchParams.start || format(subDays(new Date(), 6), "yyyy-MM-dd");
  const end = searchParams.end || format(new Date(), "yyyy-MM-dd");
  const url = `compat/wakatime/v1/users/current/summaries?${new URLSearchParams({ start, end })}`;

  const durationData = await fetchData<SummariesApiResponse>(url, true);
  if (!durationData) {
    return (
      <h3 className="text-center text-red-500">
        Error fetching dashboard stats
      </h3>
    );
  }

  const projects = makePieChartDataFromRawApiResponse(
    durationData.data,
    "projects"
  );
  return (
    <div className="my-6">
      {durationData && (
        <main className="main-dashboard space-y-5">
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

          <div className="my-5 space-y-5">
            <div className="charts-grid">
              <div className="chart-box">
                <WBarChart
                  innerRadius={34.45}
                  title="EDITORS"
                  colorNamespace="editors"
                  defaultOrientation="vertical"
                  data={makePieChartDataFromRawApiResponse(
                    durationData.data,
                    "editors"
                  )}
                  durationSubtitle="Editors used over the "
                />
              </div>
              <div className="chart-box">
                <WBarChart
                  innerRadius={34.45}
                  title="LANGUAGES"
                  data={makePieChartDataFromRawApiResponse(
                    durationData.data,
                    "languages"
                  )}
                  defaultOrientation="vertical"
                  colorNamespace="languages"
                  durationSubtitle="Languages used over the "
                />
              </div>
            </div>
            <div className="charts-grid">
              <div className="chart-box">
                <WBarChart
                  title="OPERATING SYSTEMS"
                  data={makePieChartDataFromRawApiResponse(
                    durationData.data,
                    "operating_systems"
                  )}
                  defaultOrientation="horizontal"
                  colorNamespace="operating_systems"
                  durationSubtitle="Operating Systems used over the "
                />
              </div>
              <div className="chart-box">
                <WBarChart
                  title="MACHINES"
                  data={makePieChartDataFromRawApiResponse(
                    durationData.data,
                    "machines"
                  )}
                  defaultOrientation="horizontal"
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
                <WeekdaysBarChart data={durationData.data} />
              </div>
            </div>
          </div>

          <div className="my-5">
            <div className="flex items-baseline gap-1 align-middle">
              <h1 className="text-2xl mb-4">Projects</h1>
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
