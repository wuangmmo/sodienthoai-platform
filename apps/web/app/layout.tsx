import type { Metadata } from "next";
import "./globals.css";
export const metadata:Metadata={metadataBase:new URL("https://sodienthoai.com"),title:{default:"SoDienThoai.com - Tra cứu số điện thoại",template:"%s | SoDienThoai.com"},description:"Tra cứu, xác minh và báo cáo số điện thoại."};
export default function RootLayout({children}:Readonly<{children:React.ReactNode}>){return <html lang="vi"><body>{children}</body></html>}
