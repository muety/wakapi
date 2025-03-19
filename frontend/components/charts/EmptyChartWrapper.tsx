// "use client";

// import { cn } from "@/lib/utils";
// import { LucideActivitySquare } from "lucide-react";
// import type React from "react";

// interface EmptyChartWrapperProps {
//   children: React.ReactNode;
//   emptyStateMessage?: string;
//   hasData?: boolean;
//   className?: string;
// }

// export function EmptyChartWrapper({
//   children,
//   emptyStateMessage = "No data available",
//   hasData,
//   className = "",
// }: EmptyChartWrapperProps) {
//   if (!hasData) {
//     return (
//       <div
//         className={cn(
//           "relative flex flex-col items-center justify-center h-[210px] w-full rounded-md overflow-hidden",
//           className
//         )}
//       >
//         {/* Blurred background */}
//         <div className="absolute inset-0 backdrop-blur-sm"></div>

//         {/* Content */}
//         <div className="relative z-10 flex flex-col items-center justify-center">
//           <div className="text-xl font-bold mb-2">NO CHART</div>

//           <div className="mb-2 text-gray-500">
//             <LucideActivitySquare size={32} />
//           </div>

//           <p className="text-sm text-gray-500">{emptyStateMessage}</p>
//         </div>
//       </div>
//     );
//   }

//   // If we have data, render the children (chart component)
//   return <>{children}</>;
// }

"use client";

import { LucideActivitySquare } from "lucide-react";
import type React from "react";

import { cn } from "@/lib/utils";

interface EmptyChartWrapperProps {
  children: React.ReactNode;
  emptyStateMessage?: string;
  hasData?: boolean;
  className?: string;
}

export function EmptyChartWrapper({
  children,
  emptyStateMessage = "No data available",
  hasData,
  className = "",
}: EmptyChartWrapperProps) {
  return (
    <div className={cn("relative w-full", className)}>
      {/* Always render the children */}
      {children}

      {/* Overlay the empty state only when there's no data */}
      {!hasData && (
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          {/* Blurred overlay */}
          <div className="absolute inset-0 backdrop-blur-sm rounded-md"></div>

          {/* Content */}
          <div className="relative z-10 flex flex-col items-center justify-center">
            <div className="text-xl font-bold mb-2">NO CHART</div>

            <div className="mb-2 text-gray-500">
              <LucideActivitySquare size={32} />
            </div>

            <p className="text-sm text-gray-500">{emptyStateMessage}</p>
          </div>
        </div>
      )}
    </div>
  );
}
