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
import { TooltipWithProvider } from "./tooltip-with-provider";
import { truncate } from "lodash";
import { DataItem, LeaderboardApiResponse } from "@/lib/types";

function Hireable() {
  return (
    <div
      className="border border-border rounded-md"
      style={{ color: "#afc9f2", fontSize: "0.65rem", maxWidth: "50px" }}
    >
      hireable
    </div>
  );
}

interface iProps {
  title: string;
  data: LeaderboardApiResponse;
}

function RenderLanguages({ languages }: { languages: string[] }) {
  const truncated = truncate(languages.join(", "), { length: 75 });
  const truncatedArray = truncated.split(", ");
  const last_item = truncatedArray.pop();
  const items = truncatedArray.map((item, index) => (
    <a
      key={index}
      href={`/leaderboards?languages=${item}`}
      className="text-white hover:underline gap-2"
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

export function LeaderBoardTable({ title, data: leaderboardData }: iProps) {
  const { data: rawLeaderboard, range } = leaderboardData;
  const leaderboard = rawLeaderboard.map((item) => rowMapper(item));
  return (
    <div>
      <div className="mb-2">
        <h1 className="text-4xl">{title}</h1>
      </div>
      <Table className="w-full w-100">
        <TableCaption>
          <p>
            Leaderboard for the {range.text}. {range.start_text} -{" "}
            {range.end_text}
          </p>
        </TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[70px]">Rank</TableHead>
            <TableHead className="text-left w-[300px]">Programmer</TableHead>
            <TableHead className="text-left w-[300px]">
              <div className="flex items-center gap-2">
                Hours Coded
                <TooltipWithProvider description="Total hours coded over the last 7 days from Yesterday, using default 15 minute timeout, only showing coding activity from known languages." />
              </div>
            </TableHead>
            <TableHead className="text-left w-[150px] flex items-center gap-2">
              Daily Average
              <TooltipWithProvider description="Average hours coded per day, excluding days with zero coding activity." />
            </TableHead>
            <TableHead className="text-left" style={{ maxWidth: "500px" }}>
              Languages Used
            </TableHead>
            <TableHead className="text-right w-40"></TableHead>
            <TableHead className="text-right w-24"></TableHead>
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
                className="text-right truncate"
                style={{ maxWidth: "550px" }}
              >
                <RenderLanguages languages={leader.languages} />
              </TableCell>
              <TableCell className="text-center font-semibold w-32">
                {leader.hireable && <Hireable />}
              </TableCell>
              <TableCell className="text-right font-semibold w-24"></TableCell>
            </TableRow>
          ))}
        </TableBody>
        <TableFooter></TableFooter>
      </Table>
    </div>
  );
}
