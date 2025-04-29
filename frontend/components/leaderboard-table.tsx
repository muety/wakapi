import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { DataItem, LeaderboardApiResponse } from "@/lib/types";
import { cn } from "@/lib/utils";

import { RenderLanguages } from "./render-languages";
import { TooltipWithProvider } from "./tooltip-with-provider";

export function Hireable() {
  return (
    <div
      className="rounded-md border border-border"
      style={{ color: "#afc9f2", fontSize: "0.65rem", maxWidth: "50px" }}
    >
      hireable
    </div>
  );
}

interface iProps {
  title: string;
  data: LeaderboardApiResponse;
  titleClass?: string;
  searchParams?: Record<string, any>;
}

function rowMapper(dataItem: DataItem, index: number) {
  return {
    rank: index + 1,
    programmer: dataItem.user.display_name || "Anonymous User",
    hours_coded: dataItem.running_total.human_readable_total,
    daily_average: dataItem.running_total.human_readable_daily_average,
    languages: dataItem.running_total.languages.map((l) => l.name),
    hireable: true,
  };
}

export function LeaderBoardTable({
  title,
  data: leaderboardData,
  titleClass = "",
  searchParams,
}: iProps) {
  const { data: rawLeaderboard, range } = leaderboardData;
  const users = new Set();
  const leaderboard = rawLeaderboard
    .filter((leaderData) => {
      if (users.has(leaderData.user.id)) {
        return false;
      }
      users.add(leaderData.user.id);
      return true;
    })
    .map((item, index) => rowMapper(item, index));

  const subtitle = searchParams?.language ? `- ${searchParams.language}` : "";
  return (
    <div>
      <div className="mb-2 text-left">
        <h1 className={cn("text-3xl", titleClass)}>
          {title} {subtitle}
        </h1>
      </div>
      <Table className="w-100 w-full">
        <TableCaption>
          <p>
            Leaderboard for the {range.text}. {range.start_text} -{" "}
            {range.end_text}
          </p>
        </TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[70px]">Rank</TableHead>
            <TableHead className="w-[300px] text-left">Programmer</TableHead>
            <TableHead className="w-[300px] text-left">
              <div className="flex items-center gap-2">
                Hours Coded
                <TooltipWithProvider description="Total hours coded over the last 7 days from Yesterday, using default 15 minute timeout, only showing coding activity from known languages." />
              </div>
            </TableHead>
            <TableHead className="flex w-[150px] items-center gap-2 text-left">
              Daily Average
              <TooltipWithProvider description="Average hours coded per day, excluding days with zero coding activity." />
            </TableHead>
            <TableHead className="text-left" style={{ maxWidth: "500px" }}>
              Languages Used
            </TableHead>
            <TableHead className="w-40 text-right"></TableHead>
            <TableHead className="w-24 text-right"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {leaderboard.map((leader) => (
            <TableRow key={leader.rank}>
              <TableCell>{leader.rank}</TableCell>
              <TableCell>{leader.programmer}</TableCell>
              <TableCell>{leader.hours_coded}</TableCell>
              <TableCell>{leader.daily_average}</TableCell>
              <TableCell
                className="truncate text-right"
                style={{ maxWidth: "550px" }}
              >
                <RenderLanguages languages={leader.languages} />
              </TableCell>
              <TableCell className="w-24 text-right font-semibold"></TableCell>
            </TableRow>
          ))}
        </TableBody>
        <TableFooter></TableFooter>
      </Table>
    </div>
  );
}
