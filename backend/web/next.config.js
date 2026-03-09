/** @type {import('next').NextConfig} */
const apiProxyTarget = process.env.API_PROXY_TARGET || "http://localhost:8080";

const nextConfig = {
  reactStrictMode: true,
  typescript: {
    ignoreBuildErrors: true,
  },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${apiProxyTarget}/:path*`,
      },
    ];
  },
  async redirects() {
    return [
      {
        // Redirect apex domain to canonical www domain.
        source: "/:path*",
        has: [
          {
            type: "host",
            value: "resumetailor.live",
          },
        ],
        destination: "https://www.resumetailor.live/:path*",
        permanent: true,
      },
      {
        // Redirect old Render domain to canonical www domain
        source: "/:path*",
        has: [
          {
            type: "host",
            value: "resume-tailor-web.onrender.com",
          },
        ],
        destination: "https://www.resumetailor.live/:path*",
        permanent: true,
      },
    ];
  },
};

export default nextConfig;
