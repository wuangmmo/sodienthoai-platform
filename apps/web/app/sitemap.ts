import type { MetadataRoute } from "next";
const SITE=process.env.NEXT_PUBLIC_SITE_URL||"https://sodienthoai.com";
const API=process.env.API_INTERNAL_BASE_URL||process.env.NEXT_PUBLIC_API_BASE_URL||"http://localhost:8080";
const PAGE_SIZE=40000;

export async function generateSitemaps(){
  try{
    const res=await fetch(API+"/v1/seo/sitemap/count",{next:{revalidate:3600}});
    if(!res.ok)return[{id:0}];
    const body:{count:number;page_size:number}=await res.json();
    const pages=Math.max(1,Math.ceil(body.count/body.page_size));
    return Array.from({length:pages},(_,id)=>({id}));
  }catch{return[{id:0}]}
}

async function cursorForPage(page:number){
  let after="";
  for(let i=0;i<page;i++){
    const res=await fetch(API+`/v1/seo/sitemap?limit=${PAGE_SIZE}&after=${encodeURIComponent(after)}`,{next:{revalidate:3600}});
    if(!res.ok)return null;
    const body:{next_cursor?:string}=await res.json();
    if(!body.next_cursor)return null;
    after=body.next_cursor;
  }
  return after;
}

export default async function sitemap({id}:{id:number}):Promise<MetadataRoute.Sitemap>{
  const urls:MetadataRoute.Sitemap=[];
  if(id===0)urls.push({url:SITE,changeFrequency:"daily",priority:1});
  try{
    const after=await cursorForPage(id);
    if(after===null)return urls;
    const res=await fetch(API+`/v1/seo/sitemap?limit=${PAGE_SIZE}&after=${encodeURIComponent(after)}`,{next:{revalidate:3600}});
    if(!res.ok)return urls;
    const body:{data:{e164:string;updated_at?:string}[]}=await res.json();
    for(const item of body.data)urls.push({url:SITE+"/phone/"+encodeURIComponent(item.e164),lastModified:item.updated_at?new Date(item.updated_at):undefined,changeFrequency:"weekly",priority:.6});
  }catch{}
  return urls;
}
