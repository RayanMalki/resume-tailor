import "./globals.css";
import type { ReactNode } from "react";
import type { Metadata } from "next";
import { IBM_Plex_Sans, Space_Grotesk } from "next/font/google";
import ErrorBoundary from "./components/ErrorBoundary";
import { ToastProvider } from "./components/Toast";

const grotesk = Space_Grotesk({
  subsets: ["latin"],
  variable: "--font-grotesk",
  display: "swap"
});

const plex = IBM_Plex_Sans({
  subsets: ["latin"],
  variable: "--font-plex",
  display: "swap",
  weight: ["400", "500", "600", "700"]
});

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL || "https://www.resumetailor.live";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: {
    default: "Resume Tailor",
    template: "%s | Resume Tailor"
  },
  description:
    "Tailor your resume to any job description with ATS-focused scoring and targeted improvements.",
  applicationName: "Resume Tailor",
  alternates: {
    canonical: "/"
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-image-preview": "large",
      "max-snippet": -1,
      "max-video-preview": -1
    }
  },
  openGraph: {
    title: "Resume Tailor",
    description:
      "Tailor your resume to any job description with ATS-focused scoring and targeted improvements.",
    url: "/",
    siteName: "Resume Tailor",
    type: "website"
  },
  twitter: {
    card: "summary_large_image",
    title: "Resume Tailor",
    description:
      "Tailor your resume to any job description with ATS-focused scoring and targeted improvements."
  }
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className={`${grotesk.variable} ${plex.variable}`}>
      <body className="min-h-screen font-plex">
        <a href="#main-content" className="skip-link">
          Skip to content
        </a>
        <ErrorBoundary>
          <ToastProvider>
            <div id="main-content">{children}</div>
          </ToastProvider>
        </ErrorBoundary>
      </body>
    </html>
  );
}
