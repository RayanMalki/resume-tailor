import "./globals.css";
import type { ReactNode } from "react";
import { IBM_Plex_Sans, Space_Grotesk } from "next/font/google";
import ErrorBoundary from "./components/ErrorBoundary";

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

export const metadata = {
  title: "Resume Tailor",
  description: "Tailor your resume to any job description."
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className={`${grotesk.variable} ${plex.variable}`}>
      <body className="min-h-screen font-plex">
        <a href="#main-content" className="skip-link">
          Skip to content
        </a>
        <ErrorBoundary>
          <div id="main-content">{children}</div>
        </ErrorBoundary>
      </body>
    </html>
  );
}
