import { fetchData } from "@/actions";
import { humanizeDate } from "@/lib/utils";

export interface iPluginUserAgent {
  id: string;
  value: string;
  editor: string;
  version: string;
  os: string;
  last_seen_at: string;
  is_browser_extension: boolean;
  is_desktop_app: boolean;
  created_at: string;
  cli_version: string;
  go_version: string;
}

function PluginStatusCard({ agent }: { agent: iPluginUserAgent }) {
  const status = "OK";

  return (
    <div className="plugin-status flex-col justify-center md:flex-row surface gap-5">
      <div className="flex flex-col justify-center">
        <h1 className="text-2xl">{agent.editor}</h1>
        <p className="text-sm">
          <b>Last Seen: </b>
          {humanizeDate(agent.last_seen_at)}
        </p>
        <p className="text-sm">
          <b>Version: </b>
          {agent.version} <span></span>
          with cli {agent.cli_version}
        </p>
      </div>
      <div>
        <h1 className="text-4xl font-bold text-green-500">{status}</h1>
      </div>
    </div>
  );
}

export default async function Page() {
  const userAgents = await fetchData<{ data: iPluginUserAgent[] }>(
    "compat/wakatime/v1/users/current/user-agents"
  );
  if (!userAgents) {
    return <p>Error fetching </p>;
  }

  return (
    <div
      className="flex flex-col justify-center items-center md:px-32 m-14"
      style={{ minHeight: "60vh" }}
    >
      <h1 className="text-6xl">Plugin Status</h1>
      <p className="my-5 mb-12 text-lg">
        Your plugins and their health status.
      </p>

      <div className="w-full flex flex-col justify-center mx-auto mr-12 gap-5">
        {userAgents.data.map((agent) => (
          <PluginStatusCard agent={agent} key={agent.id} />
        ))}

        {userAgents.data.length === 0 && (
          <p className="text-lg text-center">
            We have not received any plugin activity for your account. <br />{" "}
            Check your plugin setup to ensure it is working correctly, code a
            bit and come back to check again.
          </p>
        )}
      </div>
    </div>
  );
}
