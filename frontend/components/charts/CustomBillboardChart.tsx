// base css
// import "billboard.js/dist/billboard.css";
// import bb from "billboard.js";

// for ESM environment, need to import modules as:
// import bb, {line} from "billboard.js";

// var chart = bb.generate({
//   data: {
//     columns: [
// 	["data1", 30, 200, 100, 400, 150, 250],
// 	["data2", 50, 20, 10, 40, 15, 25]
//     ],
//     type: "line", // for ESM specify as: line()
//   },
//   legend: {
//     show: false
//   },
//   bindto: "#lineChart"
// });

// setTimeout(function() {
// 	chart.load({
// 		columns: [
// 			["data1", 230, 190, 300, 500, 300, 400]
// 		]
// 	});
// }, 1000);

// setTimeout(function() {
// 	chart.load({
// 		columns: [
// 			["data3", 130, 150, 200, 300, 200, 100]
// 		]
// 	});
// }, 1500);

// setTimeout(function() {
// 	chart.unload({
// 		ids: "data1"
// 	});
// }, 2000);

import React, { useEffect, useRef } from "react";
import "billboard.js/dist/billboard.css";
import bb, { ChartOptions } from "billboard.js";

const CustomBillboardChart = ({ options }: { options: ChartOptions }) => {
  console.log("options", options);
  const chartRef = useRef(null);

  useEffect(() => {
    const chart = bb.generate({
      ...options,
      bindto: chartRef.current,
    });

    // setTimeout(() => {
    //   chart.load({
    //     columns: [["data1", 230, 190, 300, 500, 300, 400]],
    //   });
    // }, 1000);

    // setTimeout(() => {
    //   chart.load({
    //     columns: [["data3", 130, 150, 200, 300, 200, 100]],
    //   });
    // }, 1500);

    // setTimeout(() => {
    //   chart.unload({
    //     ids: "data1",
    //   });
    // }, 2000);

    // Cleanup function to destroy the chart when the component unmounts
    return () => {
      chart.destroy();
    };
  }, []);

  return <div id="lineChart" ref={chartRef} style={{ maxHeight: 200 }}></div>;
};

export default CustomBillboardChart;
