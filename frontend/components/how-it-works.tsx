"use client";

import { ExternalLink } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";

import { ConfigModal } from "./config-modal";
import { FadeOnView } from "./fade-on-view";

export default function HowItWorks() {
  const [showModal, setShowModal] = useState(false);

  return (
    <FadeOnView>
      <section className="py-12 md:py-16 lg:py-20">
        <div className="container">
          <h2 className="text-3xl font-bold tracking-tight mb-12 text-center">
            How it Works
          </h2>
          <div className="grid gap-6 md:grid-cols-2 lg:gap-12">
            <div className="space-y-4">
              <ol className="list-decimal list-inside space-y-4">
                <li className="text-lg">
                  Install the relevant WakaTime plugins for your editor
                  <a
                    href="https://wakatime.com/plugins"
                    className="text-primary inline-flex items-center ml-1"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    here
                    <ExternalLink className="h-4 w-4 ml-1" />
                  </a>
                </li>
                <li className="text-lg">
                  Locate the{" "}
                  <code className="bg-muted px-1 py-0.5 rounded">
                    .wakatime.cfg
                  </code>{" "}
                  file on your computer. This is usually located in your root
                  folder. On windows you might have to show hidden files to see
                  it.
                </li>
                <li className="text-lg">
                  Update it with the API key you get in your Wakana dashboard
                  <a
                    href="https://wakana.io/settings"
                    className="text-primary inline-flex items-center ml-1"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    here
                    <ExternalLink className="h-4 w-4 ml-1" />
                  </a>
                </li>
                <li className="text-lg">
                  Update the{" "}
                  <code className="bg-muted px-1 py-0.5 rounded">api_url</code>{" "}
                  to{" "}
                  <code className="bg-muted px-1 py-0.5 rounded">
                    https://api.wakana.io/api
                  </code>
                </li>
              </ol>
              <Button onClick={() => setShowModal(true)} className="mt-4">
                View Sample Configuration
              </Button>
            </div>
            <div className="bg-muted rounded-lg p-6">
              <h3 className="text-xl font-semibold mb-4">
                After Configuration
              </h3>
              <ol className="list-decimal list-inside space-y-4">
                <li className="text-lg">
                  Open your editor and start typing something
                </li>
                <li className="text-lg">
                  Check your Wakana dashboard to see if stats show up
                </li>
                <li className="text-lg">
                  Also check the plugins section on your dashboard
                  <a
                    href="https://wakana.io/plugins/status"
                    className="text-primary inline-flex items-center ml-1"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    here
                    <ExternalLink className="h-4 w-4 ml-1" />
                  </a>
                  to see if data from any of your plugins has been collected
                </li>
              </ol>
            </div>
          </div>
        </div>

        <ConfigModal
          open={showModal}
          onOpenChange={setShowModal}
          apiKey="2df5863c-86ec-590b-a9d3-1755c28c39da"
          apiUrl="http://localhost:5509/api"
        />
      </section>
    </FadeOnView>
  );
}
