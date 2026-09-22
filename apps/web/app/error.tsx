"use client";
export default function ErrorPage({reset}:{reset:()=>void}){return <main className="wrap result"><h1>Đã có lỗi xảy ra</h1><p className="notice">Không thể tải dữ liệu lúc này. Bạn có thể thử lại mà không cần nhập lại từ đầu.</p><button className="retry" onClick={()=>reset()}>Thử lại</button></main>}
