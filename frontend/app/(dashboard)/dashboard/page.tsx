import { QuestionMarkCircledIcon } from "@radix-ui/react-icons";
import { format, subDays } from "date-fns";

import { fetchData } from "@/actions";
import { ActivityCategoriesChart } from "@/components/charts/ActivityCategoriesChart";
import { DailyCodingSummaryOverTime } from "@/components/charts/DailyCodingSummaryOverTime";
import { WeekdaysBarChart } from "@/components/charts/WeekdaysBarChart";
import { WGaugeChart } from "@/components/charts/WGaugeChart";
import { WPieChart } from "@/components/charts/WPieChart";
import { DashboardPeriodSelector } from "@/components/dashboard-period-selector";
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
import { Metadata } from "next";

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

  const durationData = await fetchData<SummariesApiResponse>(url);
  if (!durationData) {
    throw Error("Internal Server error");
  }
  const projects = durationData
    ? makePieChartDataFromRawApiResponse(durationData.data, "projects")
    : [];
  return (
    <div className="my-6">
      {durationData && (
        <main className="main-dashboard">
          <div className="m-0 my-5 text-2xl">
            <b>{durationData.cumulative_total.text}</b> <span>over the</span>{" "}
            <DashboardPeriodSelector searchParams={searchParams} />
          </div>

          <section className="charts-grid-top">
            <div className="chart-box min-h-52">
              <DailyCodingSummaryOverTime data={durationData.data} />
            </div>
            <div className="chart-box min-h-52">
              <WGaugeChart
                data={durationData.data}
                dailyAverage={durationData.daily_average}
              />
            </div>
          </section>

          <div className="my-5">
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
              <h1 className="text-2xl">Projects</h1>
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
            <div className="col-xs-12">
              <div className="my-4 flex flex-wrap" id="projects">
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
                          Check your plugin status if you have doubts about this
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
