import {NextRequest,NextResponse} from "next/server";
const base=process.env.API_INTERNAL_BASE_URL||"http://api:8080";
export async function POST(req:NextRequest){const token=req.cookies.get("control_session")?.value;if(token)await fetch(base+"/admin/auth/logout",{method:"POST",headers:{authorization:`Bearer ${token}`},cache:"no-store"}).catch(()=>null);const res=NextResponse.json({ok:true});res.cookies.set("control_session","",{httpOnly:true,secure:process.env.NODE_ENV==="production",sameSite:"lax",path:"/",maxAge:0});return res}
