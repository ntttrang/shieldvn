import { ScamAnalysisResponse } from "@/lib/api";
import { ShieldCheck, TriangleAlert, AlertCircle } from "lucide-react";

export function RiskCard({ data }: { data: ScamAnalysisResponse }) {
  const scoreConfig = {
    RED: {
      bandClass: "bg-verdict-red text-white",
      iconBg: "bg-white/20 text-white",
      icon: <TriangleAlert className="w-6 h-6" strokeWidth={2} />,
      label: "NGUY HIỂM",
      guide: "Không chuyển tiền.",
      trackClass: "bg-white/30",
      fillClass: "bg-white/95"
    },
    YELLOW: {
      bandClass: "bg-verdict-yellow text-verdict-yellow-ink",
      iconBg: "bg-verdict-yellow-ink/10 text-verdict-yellow-ink",
      icon: <AlertCircle className="w-6 h-6" strokeWidth={2} />,
      label: "CẨN THẬN",
      guide: "Tìm hiểu thêm trước khi chuyển tiền.",
      trackClass: "bg-verdict-yellow-ink/20",
      fillClass: "bg-verdict-yellow-ink/85"
    },
    GREEN: {
      bandClass: "bg-verdict-green text-white",
      iconBg: "bg-white/20 text-white",
      icon: <ShieldCheck className="w-6 h-6" strokeWidth={2} />,
      label: "AN TOÀN",
      guide: "Chưa phát hiện dấu hiệu lừa đảo.",
      trackClass: "bg-white/30",
      fillClass: "bg-white/95"
    }
  };

  const config = scoreConfig[data.risk_score] || scoreConfig.GREEN;
  const confidencePercent = Math.round(data.confidence_score * 100);

  return (
    <div className={`rounded-xl p-5 mb-4 ${config.bandClass}`}>
      <div className="flex items-center gap-2.5">
        <span className={`w-10 h-10 rounded-[10px] flex items-center justify-center shrink-0 ${config.iconBg}`}>
          {config.icon}
        </span>
        <span className="text-[26px] font-bold leading-none tracking-[-0.01em]">
          {config.label}
        </span>
      </div>
      <p className="text-[18px] font-medium my-3 leading-snug">
        {config.guide}
      </p>
      <div className="flex items-center gap-2.5">
        <span className="text-[13px] font-medium opacity-90 min-w-[96px] tabular-nums">
          Độ tin cậy {confidencePercent}%
        </span>
        <div className={`flex-1 h-2 rounded-full overflow-hidden ${config.trackClass}`}>
          <div 
            className={`h-full rounded-full transition-all duration-1000 ease-out ${config.fillClass}`} 
            style={{ width: `${confidencePercent}%` }} 
          />
        </div>
      </div>
    </div>
  );
}
