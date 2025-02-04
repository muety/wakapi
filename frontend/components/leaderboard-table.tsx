import { truncate } from "lodash";

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

import { TooltipWithProvider } from "./tooltip-with-provider";

function Hireable() {
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
}

function RenderLanguages({ languages }: { languages: string[] }) {
  const truncated = truncate(languages.join(", "), { length: 75 });
  const truncatedArray = truncated.split(", ");
  const last_item = truncatedArray.pop();
  const items = truncatedArray.map((item, index) => (
    <a
      key={index}
      href={`/leaderboards?languages=${item}`}
      className="gap-2 text-white hover:underline"
    >
      {item},
    </a>
  ));
  const totalCharacters = truncatedArray.reduce(
    (acc, item) => acc + item.length + 1,
    0
  );
  const remainingCharacters = 300 - totalCharacters;
  const lastItemText = truncate(last_item, { length: remainingCharacters });
  return (
    <div className="flex" style={{ gap: "1px" }} title={languages.join(", ")}>
      {items}
      {last_item && (
        <a
          href={`/leaderboards?languages=${last_item}`}
          className="text-white hover:underline"
        >
          {lastItemText}
        </a>
      )}
    </div>
  );
}

function rowMapper(dataItem: DataItem) {
  return {
    rank: dataItem.rank,
    programmer: dataItem.user.full_name || "Anonymous User",
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
}: iProps) {
  const { data: rawLeaderboard, range } = leaderboardData;
  const leaderboard = rawLeaderboard.map((item) => rowMapper(item));
  return (
    <div>
      <div className="mb-2">
        <h1 className={cn("text-4xl", titleClass)}>{title}</h1>
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
              <TableCell className="w-32 text-center font-semibold">
                {leader.hireable && <Hireable />}
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
