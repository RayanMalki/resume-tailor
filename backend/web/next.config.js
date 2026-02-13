/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async redirects() {
    return [
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
