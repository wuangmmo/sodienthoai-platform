import type { MetadataRoute } from "next";
const SITE="https://sodienthoai.com";
const API=process.env.API_INTERNAL_BASE_URL||process.env.NEXT_PUBLIC_API_BASE_URL||"http://localhost:8080";
export async function generateSitemaps(){
  try{const res=await fetch(API+"/v1/seo/sitemap/count",{next:{revalidate:3600}});if(!res.ok)return[{id:0}];const body:{count:number;page_size:number}=await res.json();const pages=Math.max(1,Math.ceil(body.count/body.page_size));return Array.from({length:pages},(_,id)=>({id}));}catch{return[{id:0}]}
}
export default async function sitemap({id}:{id:number}):Promise<MetadataRoute.Sitemap>{
  const urls:MetadataRoute.Sitemap=[];if(id===0)urls.push({url:SITE,changeFrequency:"daily",priority:1});
  try{const res=await fetch(API+`/v1/seo/sitemap?limit=40000&page=${id}`,{next:{revalidate:3600}});if(!res.ok)return urls;const body:{data:{e164:string;updated_at?:string}[]}=await res.json();for(const item of body.data)urls.push({url:SITE+"/phone/"+encodeURIComponent(item.e164),lastModified:item.updated_at?new Date(item.updated_at):undefined,changeFrequency:"weekly",priority:.6})}catch{}
  return urls;
}
