import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "OpsPilot AI",
  description: "Autonomous SRE incident investigation and remediation.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body className="min-h-screen flex flex-col">
        <header className="border-b border-slate-800 px-6 py-3 flex items-center justify-between shrink-0">
          <Link href="/" className="flex items-center gap-3 group">
            <span className="text-cyan-400 text-lg font-bold tracking-tight group-hover:text-cyan-300 transition-colors">
              OpsPilot AI
            </span>
            <span className="hidden sm:block text-xs font-medium text-slate-500 border border-slate-700 px-2 py-0.5 rounded-full">
              Autonomous SRE
            </span>
          </Link>
          <nav className="flex items-center gap-6 text-sm">
            <Link
              href="/analyze"
              className="text-slate-400 hover:text-white transition-colors"
            >
              Analyze
            </Link>
            <a
              href="https://github.com"
              target="_blank"
              rel="noopener noreferrer"
              className="text-slate-400 hover:text-white transition-colors"
            >
              GitHub
            </a>
          </nav>
        </header>
        <main className="flex-1">{children}</main>
      </body>
    </html>
  );
}
