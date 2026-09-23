import type { MetadataRoute } from "next";
const SITE=process.env.NEXT_PUBLIC_SITE_URL||"https://sodienthoai.com";
const production=process.env.APP_ENV==="production";
export default function robots():MetadataRoute.Robots{
 if(!production)return{rules:{userAgent:"*",disallow:"/"}};
 return{rules:{userAgent:"*",allow:"/",disallow:["/api/","/admin"]},sitemap:SITE+"/sitemap.xml",host:SITE}
}
