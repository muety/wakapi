/** @type {import("next").NextConfig} */
module.exports = {
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "avatars.githubusercontent.com",
        port: "",
        pathname: "/**",
      },
      {
        protocol: "http",
        hostname: "localhost:5509",
        port: "5509",
        pathname: "/**",
      },
      {
        protocol: "https",
        hostname: "frontend-twilight-surf-2167.fly.dev",
        port: "",
        pathname: "/**",
      },
      {
        protocol: "http",
        hostname: "*",
        port: "5509",
        pathname: "/**",
      },
    ],
  },
};
