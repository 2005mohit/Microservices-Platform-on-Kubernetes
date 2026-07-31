/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    domains: ['localhost'],
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/:path*`,
      },
      {
        source: '/ws/:path*',
        destination: `${process.env.NEXT_PUBLIC_WS_URL || 'http://localhost:8080'}/ws/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;
