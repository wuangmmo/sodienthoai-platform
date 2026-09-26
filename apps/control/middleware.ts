import {NextRequest,NextResponse} from "next/server";
export function middleware(req:NextRequest){const logged=Boolean(req.cookies.get("control_session")?.value);if(!logged&&req.nextUrl.pathname!=="/login")return NextResponse.redirect(new URL("/login",req.url));if(logged&&req.nextUrl.pathname==="/login")return NextResponse.redirect(new URL("/",req.url));return NextResponse.next()}
export const config={matcher:["/((?!api/session|_next/static|_next/image|favicon.ico).*)"]};
