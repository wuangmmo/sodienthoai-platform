"use client";

import { FormEvent, useState } from "react";

export default function LookupForm() {
  const [number,setNumber]=useState("");
  function submit(e:FormEvent){e.preventDefault();const value=number.trim();if(!value)return;window.location.href="/phone/"+encodeURIComponent(value);}
  return <form className="lookup" onSubmit={submit}>
    <label htmlFor="phone">Nhập số điện thoại cần kiểm tra</label>
    <div className="lookupRow">
      <input id="phone" inputMode="tel" autoComplete="tel" placeholder="+84 705 899 899" value={number} onChange={e=>setNumber(e.target.value)} />
      <button type="submit">Tra cứu</button>
    </div>
    <p>Hỗ trợ số quốc tế theo chuẩn E.164. Không tự đoán quốc gia để tránh nhầm số.</p>
  </form>;
}
