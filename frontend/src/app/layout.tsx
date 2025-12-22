import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "@/styles/globals.css";

/**
 * Inter font configuration
 * Subsets: Latin characters
 * Variable font with optimized loading
 */
const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});

/**
 * Application metadata configuration
 * These values are used for SEO and social sharing
 */
export const metadata: Metadata = {
  title: {
    default: "MYY Chat - AI Companion Platform",
    template: "%s | MYY Chat",
  },
  description: "Enterprise-grade AI companion platform with intelligent conversation capabilities",
  keywords: ["AI", "chat", "companion", "conversation", "artificial intelligence"],
  authors: [{ name: "MYY Chat Team" }],
  openGraph: {
    type: "website",
    locale: "zh_CN",
    url: "https://myy-chat.com",
    siteName: "MYY Chat",
    title: "MYY Chat - AI Companion Platform",
    description: "Enterprise-grade AI companion platform with intelligent conversation capabilities",
  },
  twitter: {
    card: "summary_large_image",
    title: "MYY Chat - AI Companion Platform",
    description: "Enterprise-grade AI companion platform with intelligent conversation capabilities",
  },
  robots: {
    index: true,
    follow: true,
  },
};

/**
 * Root layout component for the entire application
 *
 * This component wraps all pages and provides:
 * - Global styles via Tailwind CSS
 * - Font configuration
 * - HTML structure and metadata
 * - Future: Global state providers, theme providers
 *
 * @param props - Component props
 * @param props.children - Child components to render
 */
export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body className={`${inter.variable} font-sans antialiased`}>
        {children}
      </body>
    </html>
  );
}
