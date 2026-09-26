import assert from "node:assert/strict";import test from "node:test";import {validCsrf} from "./csrf";
function req(cookie:string,header:string){return {cookies:{get:(k:string)=>k==="control_csrf"&&cookie?{value:cookie}:undefined},headers:{get:(k:string)=>k==="x-csrf-token"?header:null}} as any}
test("csrf rejects missing token",()=>assert.equal(validCsrf(req("","")),false));
test("csrf rejects missing header",()=>assert.equal(validCsrf(req("abc","")),false));
test("csrf rejects mismatched token",()=>assert.equal(validCsrf(req("abc","abd")),false));
test("csrf accepts exact token",()=>assert.equal(validCsrf(req("abc","abc")),true));
