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
        hostname: "http://localhost:5509",
        port: "5509",
        pathname: "/**",
      },
    ],
  },
};
