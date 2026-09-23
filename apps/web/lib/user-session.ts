import {createHmac,randomBytes,timingSafeEqual} from "crypto";import {cookies} from "next/headers";
const COOKIE="sdt_user_session",MAX=60*60*24*30;
function secret(){return process.env.USER_SESSION_SECRET||process.env.ADMIN_SESSION_SECRET||""}
function sig(v:string){return createHmac("sha256",secret()).update(v).digest("base64url")}
export async function userSubject(){const raw=(await cookies()).get(COOKIE)?.value;if(!raw)return null;const [id,exp,s]=raw.split(".");if(!id||!exp||!s||Number(exp)<Math.floor(Date.now()/1000))return null;const expected=sig(id+"."+exp);if(s.length!==expected.length||!timingSafeEqual(Buffer.from(s),Buffer.from(expected)))return null;return "web:"+id}
export async function ensureUserSession(){const jar=await cookies();let raw=jar.get(COOKIE)?.value;const current=await userSubject();if(current)return current;const id=randomBytes(18).toString("base64url"),exp=String(Math.floor(Date.now()/1000)+MAX),value=id+"."+exp+"."+sig(id+"."+exp);jar.set(COOKIE,value,{httpOnly:true,secure:process.env.APP_ENV!=="development",sameSite:"lax",path:"/",maxAge:MAX});return "web:"+id}
