import type { Metadata } from "next";
import Link from "next/link";

type PhoneData={e164:string;country_code:string;calling_code:string;national_number:string;number_type?:string;verification_status:string;seo_status:string;spam_score:number;report_count:number;data_quality_score:number};
async function lookup(number:string,country=""):Promise<{data?:PhoneData;notFound?:boolean}>{
  const base=process.env.API_INTERNAL_BASE_URL||process.env.NEXT_PUBLIC_API_BASE_URL||"http://localhost:8080";
  const res=await fetch(base+"/v1/phone/"+encodeURIComponent(number)+(country?"?country="+encodeURIComponent(country):""),{cache:"no-store"});
  if(res.status===404)return{notFound:true};
  if(!res.ok)throw new Error("lookup_failed");
  return res.json();
}
function canonical(number:string){return "/phone/"+encodeURIComponent(number)}
export async function generateMetadata({params}:{params:Promise<{number:string}>}):Promise<Metadata>{
  const {number}=await params; const decoded=decodeURIComponent(number); let result;
  try{result=await lookup(decoded)}catch{return{title:"Tra cứu số điện thoại",robots:{index:false,follow:false}}}
  if(result.notFound)return{title:`Tra cứu ${decoded}`,description:`Kiểm tra thông tin số điện thoại ${decoded} trên SoDienThoai.com.`,alternates:{canonical:canonical(decoded)},robots:{index:false,follow:true}};
  const p=result.data!; const indexable=p.seo_status==="indexable"||p.seo_status==="indexed";
  return{title:`Số điện thoại ${p.e164}`,description:`Tra cứu thông tin, trạng thái xác minh và báo cáo của số điện thoại ${p.e164}.`,alternates:{canonical:canonical(p.e164)},robots:{index:indexable,follow:true}};
}
export default async function PhonePage({params,searchParams}:{params:Promise<{number:string}>;searchParams:Promise<{country?:string}>}){
  const {number}=await params; const {country}=await searchParams; let result;
  try{result=await lookup(decodeURIComponent(number),country||"")}catch{return <main className="wrap result"><h1>Không thể tra cứu lúc này</h1><p className="notice">Dịch vụ đang tạm thời không khả dụng. Vui lòng thử lại sau.</p><Link className="back" href="/">← Tra cứu số khác</Link></main>}
  if(result.notFound)return <main className="wrap result"><h1>{decodeURIComponent(number)}</h1><p className="notice">Số điện thoại này chưa được định danh trong dữ liệu SoDienThoai.com.</p><Link className="back" href="/">← Tra cứu số khác</Link></main>;
  const p=result.data!;
  return <main className="wrap result"><h1>{p.e164}</h1><div className="grid">
    <div className="item"><b>Quốc gia</b>{p.country_code}</div><div className="item"><b>Loại số</b>{p.number_type||"Chưa xác định"}</div>
    <div className="item"><b>Xác minh</b>{p.verification_status}</div><div className="item"><b>Báo cáo</b>{p.report_count}</div>
    <div className="item"><b>Điểm spam</b>{p.spam_score}</div><div className="item"><b>Chất lượng dữ liệu</b>{p.data_quality_score}/100</div>
  </div><Link className="back" href="/">← Tra cứu số khác</Link></main>;
}
