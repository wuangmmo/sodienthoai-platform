import {NextRequest,NextResponse} from "next/server";
import {clearAdminSession} from "../../../lib/admin-session";
export async function POST(req:NextRequest){await clearAdminSession();return NextResponse.redirect(new URL("/admin/login",req.url),303)}
