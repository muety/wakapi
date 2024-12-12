import {
  AwaitedReactNode,
  JSXElementConstructor,
  ReactElement,
  ReactNode,
  ReactPortal,
} from "react";

import { SimpleFooter } from "@/components/simple-footer";

export default async function MainContentWrapper(props: {
  children:
    | ReactElement<unknown, string | JSXElementConstructor<unknown>>
    | Iterable<ReactNode>
    | ReactPortal
    | Promise<AwaitedReactNode>
    | null
    | undefined;
}) {
  return (
    <div className="main-bg flex max-h-screen w-full flex-col">
      <main className="min-h-screen min-w-0 flex-1 md:ml-52 lg:flex">
        <div className="flex h-full min-w-0 flex-1 flex-col justify-between overflow-y-auto lg:order-last">
          <div className="mx-auto w-full">
            <div className="flex-col md:flex">
              <main className="min-h-full px-5" style={{ minHeight: "50vh" }}>
                {props.children}
              </main>
            </div>
          </div>
          <SimpleFooter />
        </div>
      </main>
    </div>
  );
}
