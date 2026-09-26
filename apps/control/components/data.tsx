"use client";import {useEffect,useState} from "react";
export function useControl<T>(path:string){const [data,setData]=useState<T|null>(null),[error,setError]=useState("");useEffect(()=>{fetch("/api/control/"+path).then(async r=>{if(!r.ok)throw new Error(String(r.status));return r.json()}).then(setData).catch(()=>setError("Không thể tải dữ liệu."))},[path]);return {data,error}}
export function State({error}:{error:string}){return <div className="card muted">{error||"Đang tải dữ liệu…"}</div>}
