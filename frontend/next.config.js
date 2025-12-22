/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,

  // Enable SWC minification for better performance
  swcMinify: true,

  // Configure image domains if needed
  images: {
    domains: [],
    formats: ['image/avif', 'image/webp'],
  },

  // Environment variables that should be available on the client
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  },

  // Experimental features
  experimental: {
    // Enable type-checking in the build process
    typedRoutes: true,
  },

  // Webpack configuration
  webpack: (config) => {
    return config;
  },
};

module.exports = nextConfig;
