import { ScamAnalysisResponse } from "@/lib/api";
import { Check } from "lucide-react";

export function EvidenceList({ data }: { data: ScamAnalysisResponse }) {
  const hasEntities = data.extracted_entities && 
    (data.extracted_entities.bank_account || data.extracted_entities.phone_number || data.extracted_entities.url);

  const getDotClass = (score: string) => {
    switch (score) {
      case "RED": return "bg-verdict-red-dot";
      case "YELLOW": return "bg-verdict-yellow-dot";
      case "GREEN": return "bg-verdict-green-dot";
      default: return "bg-verdict-green-dot";
    }
  };

  const getCheckClass = (score: string) => {
    switch (score) {
      case "RED": return "bg-verdict-red-bg text-verdict-red";
      case "YELLOW": return "bg-verdict-yellow-bg text-yellow-600";
      case "GREEN": return "bg-verdict-green-bg text-verdict-green";
      default: return "bg-verdict-green-bg text-verdict-green";
    }
  };

  const dotClass = getDotClass(data.risk_score);
  const checkClass = getCheckClass(data.risk_score);

  return (
    <div className="space-y-4">
      {/* Evidence */}
      {data.evidence && data.evidence.length > 0 && (
        <section>
          <h3 className="text-[17px] font-bold my-2 text-ink">
            {data.risk_score === "GREEN" ? "Đã kiểm tra gì" : "Tại sao lại vậy?"}
          </h3>
          <div className="bg-surface border border-line rounded-xl p-3.5">
            <ul className="m-0 p-0 list-none">
              {data.evidence.map((item, idx) => (
                <li key={idx} className="flex gap-2.5 py-2 text-[16px] leading-[1.4] border-b border-dashed border-line last:border-0 text-ink">
                  <span className={`w-2 h-2 rounded-full mt-[9px] shrink-0 ${dotClass}`} />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </div>
        </section>
      )}

      {/* Recommendations */}
      {data.recommendations && data.recommendations.length > 0 && (
        <section>
          <h3 className="text-[17px] font-bold my-2 text-ink">Bạn nên làm gì</h3>
          <div className="bg-surface border border-line rounded-xl p-3.5">
            <ul className="m-0 p-0 list-none">
              {data.recommendations.map((rec, idx) => {
                // Emphasize the first word if it's all caps (like DỪNG)
                const parts = rec.split(" — ");
                return (
                  <li key={idx} className="flex gap-[11px] items-start py-2 text-[16px] leading-[1.4] text-ink">
                    <span className={`w-6 h-6 rounded-full shrink-0 flex items-center justify-center mt-px ${checkClass}`}>
                      <Check className="w-4 h-4" strokeWidth={3} />
                    </span>
                    <span>
                      {parts.length > 1 ? (
                        <>
                          <strong>{parts[0]}</strong> — {parts.slice(1).join(" — ")}
                        </>
                      ) : (
                        <span>{rec}</span>
                      )}
                    </span>
                  </li>
                );
              })}
            </ul>
          </div>
        </section>
      )}

      {/* Extracted Entities */}
      {hasEntities && (
        <section>
          <h3 className="text-[17px] font-bold my-2 text-ink">Thông tin được trích xuất</h3>
          <div className="bg-surface border border-line rounded-xl p-3.5 flex flex-col gap-3">
            {data.extracted_entities.bank_account && (
              <div className="flex justify-between items-center text-[15px] border-b border-dashed border-line pb-2">
                <span className="text-muted font-medium">Số tài khoản</span>
                <span className="font-bold text-ink bg-slate-100 px-2 py-0.5 rounded tabular-nums">{data.extracted_entities.bank_account}</span>
              </div>
            )}
            {data.extracted_entities.phone_number && (
              <div className="flex justify-between items-center text-[15px] border-b border-dashed border-line pb-2">
                <span className="text-muted font-medium">Số điện thoại</span>
                <span className="font-bold text-ink bg-slate-100 px-2 py-0.5 rounded tabular-nums">{data.extracted_entities.phone_number}</span>
              </div>
            )}
            {data.extracted_entities.url && (
              <div className="flex flex-col gap-1 text-[15px] pt-1">
                <span className="text-muted font-medium">Đường link</span>
                <span className="font-medium text-brand break-all">{data.extracted_entities.url}</span>
              </div>
            )}
          </div>
        </section>
      )}
    </div>
  );
}
