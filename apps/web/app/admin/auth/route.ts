import {NextRequest,NextResponse} from "next/server";
import {timingSafeEqual} from "crypto";
import {setAdminSession} from "../../../lib/admin-session";
export async function POST(req:NextRequest){const form=await req.formData();const supplied=String(form.get("password")||"");const expected=process.env.ADMIN_LOGIN_PASSWORD||"";const a=Buffer.from(supplied),b=Buffer.from(expected);if(!expected||a.length!==b.length||!timingSafeEqual(a,b))return NextResponse.redirect(new URL("/admin/login?error=1",req.url),303);await setAdminSession();return NextResponse.redirect(new URL("/admin",req.url),303)}
