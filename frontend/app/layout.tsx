import type { Metadata } from "next";
import { Be_Vietnam_Pro } from "next/font/google";
import "./globals.css";

const beVietnamPro = Be_Vietnam_Pro({ 
  subsets: ["latin", "vietnamese"],
  weight: ['400', '500', '600', '700'],
});

export const metadata: Metadata = {
  title: "ShieldVN | Scam Detection",
  description: "Privacy-first scam detection platform tailored for Vietnamese users.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="vi">
      <body className={`${beVietnamPro.className} bg-bg text-ink antialiased min-h-screen`}>
        {children}
      </body>
    </html>
  );
}
