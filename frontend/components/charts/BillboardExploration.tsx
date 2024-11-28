// "use client";

// //@ts-ignore
// import "billboard.js/dist/billboard.css";

// import { areaStep, ChartOptions } from "billboard.js";

// import { SAMPLE_COLORS } from "@/lib/constants";
// import { SummariesResponse } from "@/lib/types";
// import {
//   brightenHexColor,
//   normalizeChartData,
//   prepareDailyCodingData,
// } from "@/lib/utils";

// import CustomBillboardChart from "./CustomBillboardChart";

// // for ESM environment, need to import modules as:
// // import bb, {step, areaStep} from "billboard.js";

// //   var chart2 = bb.generate({
// //     title: {
// //       text: "step-before"
// //     },
// //     data: {
// //       columns: [
// //       ["data1", 300, 350, 300, 20, 240, 100],
// //       ["data2", 130, 100, 140, 200, 150, 50]
// //       ],
// //       types: {
// //         data1: "step", // for ESM specify as: step()
// //         data2: "area-step", // for ESM specify as: areaStep()
// //       }
// //     },
// //     line: {
// //       step: {
// //         type: "step-before",
// //         tooltipMatch: true
// //       }
// //     },
// //     bindto: "#stepChart_2"
// //   });

// //   var chart3 = bb.generate({
// //     title: {
// //       text: "step-after"
// //     },
// //     data: {
// //       columns: [
// //       ["data1", 300, 350, 300, 20, 240, 100],
// //       ["data2", 130, 100, 140, 200, 150, 50]
// //       ],
// //       types: {
// //         data1: "step", // for ESM specify as: step()
// //         data2: "area-step", // for ESM specify as: areaStep()
// //       }
// //     },
// //     line: {
// //       step: {
// //         type: "step-after",
// //         tooltipMatch: true
// //       }
// //     },
// //     bindto: "#stepChart_3"
// //   });

// // const data = {
// //   data: {
// //     columns: [["data", 91.4]],
// //     type: "gauge", // for ESM specify as: gauge()
// //   },
// //   gauge: {},
// //   color: {
// //     pattern: ["#FF0000", "#F97600", "#F6C600", "#60B044"],
// //     threshold: {
// //       values: [30, 60, 90, 100],
// //     },
// //   },
// //   size: {
// //     height: 180,
// //   },
// //   bindto: "#gaugeChart",
// // };

// // const CHART_DATA = {
// //   columns: [["data", 91.4]],
// //   type: "gauge",
// //   color: {
// //     pattern: ["#FF0000", "#F97600", "#F6C600", "#60B044"],
// //     threshold: {
// //       values: [30, 60, 90, 100],
// //     },
// //   },
// // };

// export interface iProps {
//   data: SummariesResponse[];
// }

// export function BBDailyCodingActivity({ data }: iProps) {
//   const key_values: Record<string, any> = {};
//   const names = [];

//   for (const raw of data) {
//     names.push(raw.range.date);
//     for (const project of raw.projects) {
//       if (key_values[project.name]) {
//         key_values[project.name].push(project.total_seconds);
//       } else {
//         key_values[project.name] = [project.total_seconds];
//       }
//     }
//   }
//   // const unNormalized = data.map(prepareDailyCodingData);
//   const [, uniqueProjects] = normalizeChartData(
//     data.map(prepareDailyCodingData)
//   );

//   const customUniqueProjects: Record<string, string> = uniqueProjects.reduce(
//     (prev, cur, index) => ({
//       [cur]: brightenHexColor(SAMPLE_COLORS[index], 0.3),
//       ...prev,
//     }),
//     {}
//   );
//   const StepBeforeChartData: ChartOptions = {
//     // title: {
//     //   text: "default",
//     // },
//     data: {
//       columns: Object.entries(key_values).map(([key, values]) => [
//         key,
//         ...values,
//       ]),
//       types: Object.keys(customUniqueProjects).reduce(
//         (prev, cur) => ({ [cur]: areaStep(), ...prev }),
//         {}
//       ),
//       colors: customUniqueProjects,
//     },
//     axis: {
//       y: {
//         show: false,
//       },
//     },
//     // legends: Object.keys(customUniqueProjects).reduce(
//     //   (prev, cur) => ({ [cur]: { show: false }, ...prev }),
//     //   {}
//     // ),
//     legend: {
//       show: false,
//     },
//     point: {
//       show: false,
//     },

//     // line: {
//     //   step: {
//     //     type: "step-after",
//     //     tooltipMatch: true,
//     //   },
//     // },
//     bindto: "#stepChart_1",
//   };
//   return (
//     // <div>
//     <CustomBillboardChart options={StepBeforeChartData as any} />
//     // </div>
//   );
// }
