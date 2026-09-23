import {NextRequest,NextResponse} from "next/server";
import {timingSafeEqual} from "crypto";
import {setAdminSession} from "../../../lib/admin-session";
import {adminLoginAllowed,adminLoginFailed,adminLoginSucceeded} from "../../../lib/admin-login-limit";
export async function POST(req:NextRequest){if(!adminLoginAllowed(req))return NextResponse.redirect(new URL("/admin/login?error=rate",req.url),303);const form=await req.formData();const supplied=String(form.get("password")||"");const expected=process.env.ADMIN_LOGIN_PASSWORD||"";const a=Buffer.from(supplied),b=Buffer.from(expected);if(!expected||a.length!==b.length||!timingSafeEqual(a,b)){adminLoginFailed(req);return NextResponse.redirect(new URL("/admin/login?error=1",req.url),303)}adminLoginSucceeded(req);await setAdminSession();return NextResponse.redirect(new URL("/admin",req.url),303)}
