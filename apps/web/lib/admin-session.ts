import {createHmac,timingSafeEqual} from "crypto";
import {cookies} from "next/headers";

const COOKIE="sdt_admin_session";
const MAX_AGE=60*60*8;

function secret(){const v=process.env.ADMIN_SESSION_SECRET;if(!v)throw new Error("ADMIN_SESSION_SECRET is required");return v}
function sign(payload:string){return createHmac("sha256",secret()).update(payload).digest("base64url")}
export function createAdminSession(){const payload=Buffer.from(JSON.stringify({exp:Date.now()+MAX_AGE*1000})).toString("base64url");return payload+"."+sign(payload)}
export function validAdminSession(value?:string){if(!value)return false;const [payload,sig]=value.split(".");if(!payload||!sig)return false;const expected=sign(payload);const a=Buffer.from(sig),b=Buffer.from(expected);if(a.length!==b.length||!timingSafeEqual(a,b))return false;try{return JSON.parse(Buffer.from(payload,"base64url").toString()).exp>Date.now()}catch{return false}}
export async function isAdmin(){return validAdminSession((await cookies()).get(COOKIE)?.value)}
export async function setAdminSession(){(await cookies()).set(COOKIE,createAdminSession(),{httpOnly:true,secure:true,sameSite:"strict",path:"/admin",maxAge:MAX_AGE})}
export async function clearAdminSession(){(await cookies()).set(COOKIE,"",{httpOnly:true,secure:true,sameSite:"strict",path:"/admin",maxAge:0})}
