export * from "./goal";

export const Editors = [
  {
    name: "VScode",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#3279c6",
  },
  {
    name: "Vim",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#2da02d",
  },
  {
    name: "IntelliJ",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#f1e05a",
  },
];

export const OperatingSystems = [
  {
    name: "Mac",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#3279c6",
  },
  {
    name: "Ubuntu",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#2da02d",
  },
];

export const Machines = [
  {
    name: "v8.local",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#3279c6",
  },
  {
    name: "Neo's iMac",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#2da02d",
  },
];

const Languages = [
  {
    name: "Typescript",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#3279c6",
  },
  {
    name: "Golang",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#2da02d",
  },
  {
    name: "JavaScript",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#f1e05a",
  },
  {
    name: "Rust",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#d62829",
  },
  {
    name: "Vue.js",
    value: Math.floor(Math.random() * 60 * 24),
    color: "#3279c6",
  },
  {
    name: "CSS",
    value: 556,
    color: "#563d7c",
  },
  {
    name: "Bash",
    value: 256,
    color: "#2da02d",
  },
  {
    name: "Markdown",
    color: "#2da02d",
    value: 100,
  },
];

export const chartData = {
  Editors,
  Languages,
  Machines,
  OperatingSystems,
};
