"use client";

import { useState } from "react";
import { analyzeText, ScamAnalysisResponse } from "@/lib/api";
import { RiskCard } from "@/components/risk-card";
import { EvidenceList } from "@/components/evidence-list";
import { HelpCircle, Lock, ArrowRight, ArrowLeft, ImagePlus, Loader2 } from "lucide-react";

export default function Home() {
  const [text, setText] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ScamAnalysisResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;

    setLoading(true);
    setError(null);
    try {
      const data = await analyzeText(text);
      setResult(data);
    } catch (err: any) {
      setError(err.message || "Không kết nối được với máy chủ. Kiểm tra mạng và thử lại nhé.");
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    setResult(null);
    setText("");
    setError(null);
  };

  return (
    <main className="flex flex-col items-center pt-4 md:pt-7 pb-16 px-4">
      <div className="w-full max-w-[480px] bg-bg md:bg-surface md:shadow-app md:border border-line rounded-[24px] overflow-hidden min-h-[600px] flex flex-col relative">
        
        {/* Header */}
        <header className="flex items-center justify-between px-1 py-3.5 mx-3">
          {result ? (
            <div className="flex items-center gap-3">
              <button 
                onClick={handleReset} 
                className="w-10 h-10 rounded-[10px] border border-line bg-white flex items-center justify-center text-muted hover:text-brand hover:border-brand transition-colors"
                aria-label="Quay lại"
              >
                <ArrowLeft className="w-5 h-5" />
              </button>
              <div className="flex items-center gap-2.5 font-bold text-[18px] tracking-[-0.01em]">
                <div className="w-8 h-8 rounded-lg bg-brand flex items-center justify-center text-white shrink-0">
                  <ShieldLogo />
                </div>
                ShieldVN
              </div>
            </div>
          ) : (
            <>
              <div className="flex items-center gap-2.5 font-bold text-[18px] tracking-[-0.01em]">
                <div className="w-8 h-8 rounded-lg bg-brand flex items-center justify-center text-white shrink-0">
                  <ShieldLogo />
                </div>
                ShieldVN
              </div>
              <button 
                className="w-10 h-10 rounded-[10px] border border-line bg-white flex items-center justify-center text-muted hover:text-brand hover:border-brand transition-colors"
                aria-label="Trợ giúp"
              >
                <HelpCircle className="w-5 h-5" />
              </button>
            </>
          )}
        </header>

        {/* Content Area */}
        <div className="px-4 pb-7 flex-1 flex flex-col">
          {!result && !loading && (
            <div className="flex-1 flex flex-col animate-in fade-in duration-300">
              <h1 className="text-[24px] font-bold leading-[1.2] mt-0.5 mb-1.5 tracking-[-0.02em]">
                Kiểm tra lừa đảo trong 10 giây
              </h1>
              <p className="text-muted text-[16px] leading-[1.45] mb-[18px]">
                Tải ảnh hoặc dán tin nhắn đáng ngờ để nhận kết quả ngay.
              </p>

              <form onSubmit={handleSubmit} className="flex flex-col flex-1">
                <div 
                  className="border-2 border-dashed border-[#BBD4D2] rounded-xl bg-surface p-[26px_16px] text-center cursor-pointer transition-all duration-150 hover:border-brand hover:bg-brand-tint"
                  role="button"
                  tabIndex={0}
                  aria-label="Tải ảnh màn hình"
                >
                  <div className="w-[46px] h-[46px] rounded-xl bg-brand-tint text-brand flex items-center justify-center mx-auto mb-3">
                    <ImagePlus className="w-6 h-6" />
                  </div>
                  <p className="text-[17px] font-semibold m-0 mb-0.5">Tải ảnh màn hình</p>
                  <p className="text-[14px] text-muted m-0 mb-2.5">Kéo thả hoặc chạm để chọn</p>
                  <p className="text-[12px] text-[#94A3B8] tracking-[0.02em]">JPG · PNG · tối đa 5MB</p>
                  <p className="text-xs text-muted mt-2 italic">(Tính năng đang phát triển)</p>
                </div>

                <div className="flex items-center gap-3 text-[#94A3B8] text-[12px] font-semibold tracking-[0.08em] my-4 before:content-[''] before:h-px before:bg-line before:flex-1 after:content-[''] after:h-px after:bg-line after:flex-1">
                  HOẶC
                </div>

                <div className="flex-1 min-h-[108px]">
                  <textarea
                    value={text}
                    onChange={(e) => setText(e.target.value)}
                    placeholder="Dán tin nhắn / link đáng ngờ vào đây...&#10;VD: Tuyển CTV nạp tiền chốt đơn, hoa hồng cao..."
                    className="w-full min-h-[108px] h-full resize-y border border-line rounded-[10px] p-[12px_14px] text-[16px] text-ink bg-surface outline-none transition-all duration-150 focus:border-brand focus:ring-[3px] focus:ring-brand/15 placeholder:text-[#94A3B8]"
                  />
                </div>

                <div className="flex gap-[10px] items-start bg-brand-tint border border-[#BFE9E5] rounded-[10px] p-[11px_13px] mt-1">
                  <span className="text-brand mt-px"><Lock className="w-4 h-4" /></span>
                  <p className="m-0 text-[13.5px] leading-[1.4] text-[#005050]">
                    Thông tin nhạy cảm (SĐT, STK, CCCD) tự bị <strong>CHE</strong> trước khi AI phân tích.
                  </p>
                </div>

                <div className="mt-[18px]">
                  <button
                    type="submit"
                    disabled={!text.trim()}
                    className="flex items-center justify-center gap-[9px] w-full min-h-[52px] border-none rounded-[10px] text-[17px] font-semibold cursor-pointer transition-all duration-150 bg-brand text-white hover:bg-brand-600 disabled:bg-[#B7C9C7] disabled:text-white disabled:cursor-not-allowed"
                  >
                    Kiểm tra ngay
                    <ArrowRight className="w-5 h-5" />
                  </button>
                </div>
                <p className="text-center text-[#94A3B8] text-[13px] mt-[14px]">
                  Đã kiểm tra 12.480 tin nhắn
                </p>
              </form>
            </div>
          )}

          {loading && (
            <div className="flex-1 flex flex-col animate-in fade-in duration-300">
              <h1 className="text-[24px] font-bold leading-[1.2] mt-0.5 mb-1.5 tracking-[-0.02em]">
                Kiểm tra lừa đảo trong 10 giây
              </h1>
              
              <div className="text-center p-[34px_16px_30px] bg-surface border border-line rounded-xl mt-1.5 shadow-sm">
                <div className="relative w-[104px] h-[104px] mx-auto mb-[18px]">
                  <div className="absolute inset-0 rounded-full border-[1.5px] border-brand/35"></div>
                  <div className="absolute inset-[22px] rounded-full border-[1.5px] border-brand/25"></div>
                  
                  {/* Radar sweep simulation */}
                  <div className="absolute inset-0 rounded-full border-[1.5px] border-brand border-r-transparent border-b-transparent animate-spin"></div>
                  
                  <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-10 h-10 bg-brand rounded-[10px] flex items-center justify-center text-white shadow-[0_0_0_6px_rgba(0,107,107,0.12)]">
                    <ShieldLogo />
                  </div>
                </div>
                
                <p className="text-[18px] font-semibold m-0 mb-[14px] text-ink">Đang phân tích...</p>
                
                <div className="flex flex-col gap-[9px] text-left max-w-[300px] mx-auto">
                  <div className="flex items-center gap-[10px] text-[15px] text-ink">
                    <span className="w-[22px] h-[22px] rounded-full bg-verdict-green text-white flex items-center justify-center shrink-0">
                      <svg className="w-3 h-3" strokeWidth={3} viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M20 6 9 17l-5-5"/></svg>
                    </span>
                    Đã che thông tin nhạy cảm
                  </div>
                  <div className="flex items-center gap-[10px] text-[15px] text-ink">
                    <span className="w-[22px] h-[22px] rounded-full bg-verdict-green text-white flex items-center justify-center shrink-0">
                      <svg className="w-3 h-3" strokeWidth={3} viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M20 6 9 17l-5-5"/></svg>
                    </span>
                    Đã đối chiếu danh sách lừa đảo
                  </div>
                  <div className="flex items-center gap-[10px] text-[15px] text-brand font-semibold">
                    <span className="w-[22px] h-[22px] rounded-full bg-brand text-white flex items-center justify-center shrink-0 animate-pulse">
                      <div className="w-1.5 h-1.5 bg-white rounded-full"></div>
                    </span>
                    AI đang đọc văn bản...
                  </div>
                </div>
              </div>
            </div>
          )}

          {error && !loading && (
            <div className="flex-1 flex flex-col animate-in fade-in duration-300">
               <h1 className="text-[24px] font-bold leading-[1.2] mt-0.5 mb-1.5 tracking-[-0.02em]">
                Kiểm tra lừa đảo trong 10 giây
              </h1>

              <div className="flex gap-[11px] items-start bg-verdict-red-bg border border-verdict-red-border rounded-[10px] p-[13px_14px] mt-4 shadow-sm" role="alert">
                <span className="text-verdict-red mt-px"><HelpCircle className="w-5 h-5" /></span>
                <div className="flex-1">
                  <p className="m-0 mb-3 text-[15px] text-[#7F1D1D] leading-[1.4]">
                    {error}
                  </p>
                  <button 
                    onClick={handleSubmit}
                    className="min-h-[44px] px-4 font-semibold rounded-lg bg-verdict-red text-white text-[15px] hover:bg-[#B91C1C] transition-colors"
                  >
                    Thử lại
                  </button>
                </div>
              </div>
            </div>
          )}

          {result && (
            <div className="flex-1 flex flex-col animate-in slide-in-from-right-4 duration-300">
              <RiskCard data={result} />
              <EvidenceList data={result} />
              
              <div className="flex gap-2.5 mt-6">
                <button className="flex-1 min-h-[46px] bg-brand text-white font-semibold text-[15px] rounded-[10px] flex items-center justify-center gap-2 hover:bg-brand-600 transition-colors shadow-sm">
                  Tạo thiệp cảnh báo
                </button>
                <button 
                  onClick={handleReset}
                  className="flex-1 min-h-[46px] bg-surface text-ink font-semibold text-[15px] rounded-[10px] border border-line flex items-center justify-center gap-2 hover:border-brand hover:text-brand transition-colors shadow-sm"
                >
                  Kiểm tra tiếp
                </button>
              </div>
              <span className="block text-center text-muted text-[14px] mt-4">
                Kết quả sai? <a href="#" className="text-brand font-medium no-underline hover:underline">Báo cáo lại</a>
              </span>
            </div>
          )}
        </div>
      </div>
    </main>
  );
}

function ShieldLogo() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
    </svg>
  );
}
