"use client";
import {FormEvent,useState} from "react";
export default function LookupForm(){
 const[number,setNumber]=useState("");
 function submit(e:FormEvent){e.preventDefault();const value=number.trim();if(!value)return;window.location.href="/phone/"+encodeURIComponent(value)+"?country=VN"}
 return <form className="lookup" onSubmit={submit}><label htmlFor="phone">Nhập số điện thoại cần kiểm tra</label><div className="lookupRow"><input id="phone" inputMode="tel" autoComplete="tel" placeholder="0705 899 899 hoặc +84 705 899 899" value={number} onChange={e=>setNumber(e.target.value)}/><button type="submit">Tra cứu</button></div><p>Mặc định nhận dạng số nội địa Việt Nam. Số quốc tế nhập kèm mã quốc gia, ví dụ +1, +81, +65.</p></form>
}
