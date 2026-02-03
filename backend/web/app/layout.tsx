import "./globals.css";
import type { ReactNode } from "react";
import VisitBeacon from "./components/VisitBeacon";

export const metadata = {
  title: "Resume Tailor",
  description: "Tailor your resume to any job description."
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen">
        <VisitBeacon />
        {children}
      </body>
    </html>
  );
}
