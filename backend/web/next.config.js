/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async redirects() {
    return [
      {
        // Redirect old Render domain to new custom domain
        source: "/:path*",
        has: [
          {
            type: "host",
            value: "resume-tailor-web.onrender.com",
          },
        ],
        destination: "https://resumetailor.live/:path*",
        permanent: true, // 301 redirect — tells search engines the move is permanent
      },
      {
        // Redirect www to non-www for consistency
        source: "/:path*",
        has: [
          {
            type: "host",
            value: "www.resumetailor.live",
          },
        ],
        destination: "https://resumetailor.live/:path*",
        permanent: true,
      },
    ];
  },
};

export default nextConfig;
