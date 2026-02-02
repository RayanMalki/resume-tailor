import "./globals.css";
import type { ReactNode } from "react";

export const metadata = {
  title: "Resume Tailor",
  description: "Tailor your resume to any job description."
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen">
        {children}
      </body>
    </html>
  );
}
