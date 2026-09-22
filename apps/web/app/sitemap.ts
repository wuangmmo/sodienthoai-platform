import type { MetadataRoute } from "next";
type SitemapItem={e164:string;updated_at?:string};
export default async function sitemap():Promise<MetadataRoute.Sitemap>{
  const base=process.env.API_INTERNAL_BASE_URL||process.env.NEXT_PUBLIC_API_BASE_URL||"http://localhost:8080";
  const urls:MetadataRoute.Sitemap=[{url:"https://sodienthoai.com",changeFrequency:"daily",priority:1}];
  try{
    const res=await fetch(base+"/v1/seo/sitemap?limit=50000",{next:{revalidate:3600}});
    if(!res.ok)return urls;
    const body:{data:SitemapItem[]}=await res.json();
    for(const item of body.data)urls.push({url:"https://sodienthoai.com/phone/"+encodeURIComponent(item.e164),lastModified:item.updated_at?new Date(item.updated_at):undefined,changeFrequency:"weekly",priority:.6});
  }catch{}
  return urls;
}
