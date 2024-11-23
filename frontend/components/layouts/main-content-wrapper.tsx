import { SimpleFooter } from "@/components/simple-footer";
import {
  ReactElement,
  JSXElementConstructor,
  ReactNode,
  ReactPortal,
  AwaitedReactNode,
} from "react";

export default async function MainContentWrapper(props: {
  children:
    | ReactElement<any, string | JSXElementConstructor<any>>
    | Iterable<ReactNode>
    | ReactPortal
    | Promise<AwaitedReactNode>
    | null
    | undefined;
}) {
  return (
    <div className="flex max-h-screen w-full flex-col main-bg">
      <main className="min-h-screen min-w-0 flex-1 md:ml-52 lg:flex">
        <div className="flex min-h-100 h-full min-w-0 flex-1 flex-col justify-between overflow-y-auto lg:order-last">
          <div className="mx-auto w-full">
            <div className="flex-col md:flex">
              <main
                className="min-h-full min-h-100 px-5"
                style={{ minHeight: "50vh" }}
              >
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
