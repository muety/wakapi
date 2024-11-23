// import React from "react";
// import {
//   BarChart,
//   Bar,
//   Line,
//   XAxis,
//   YAxis,
//   CartesianGrid,
//   Tooltip,
//   Legend,
// } from "recharts";

// const data = [
//   {
//     name: "Jan",
//     commits: 40,
//     issues: 10,
//     reviews: 20,
//   },
//   {
//     name: "Feb",
//     commits: 50,
//     issues: 5,
//     reviews: 30,
//   },
//   {
//     name: "Mar",
//     commits: 35,
//     issues: 15,
//     reviews: 25,
//   },
//   {
//     name: "Apr",
//     commits: 60,
//     issues: 8,
//     reviews: 40,
//   },
// ];

// const StackedBarChartWithLineChart = () => {
//   return (
//     <BarChart width={700} height={300} data={data}>
//       <CartesianGrid strokeDasharray="5 5" />
//       <XAxis dataKey="name" tickLine={false} />
//       <YAxis />
//       <Tooltip />
//       <Legend verticalAlign="top" height={36} />
//       <Bar stackId="a" dataKey="commits" fill="#8884d8" name="Commits" />
//       <Bar stackId="a" dataKey="issues" fill="#82ca9d" name="Issues" />
//       <Bar stackId="a" dataKey="reviews" fill="#ffc107" name="Reviews" />
//       <Line
//         type="monotone"
//         dataKey="commits"
//         stroke="#8884d8"
//         strokeWidth={2}
//       />
//     </BarChart>
//   );
// };

// export default StackedBarChartWithLineChart;
