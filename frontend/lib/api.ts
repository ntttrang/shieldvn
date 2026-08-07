export interface ExtractedEntities {
  bank_account?: string;
  phone_number?: string;
  url?: string;
}

export interface ScamAnalysisResponse {
  risk_score: "GREEN" | "YELLOW" | "RED";
  confidence_score: number;
  detected_patterns: string[];
  evidence: string[];
  recommendations: string[];
  extracted_entities: ExtractedEntities;
}

export async function analyze(text: string, image?: File | null): Promise<ScamAnalysisResponse> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  
  const formData = new FormData();
  if (text) formData.append("text_prompt", text);
  if (image) formData.append("image", image);

  const response = await fetch(`${apiUrl}/api/v1/analyze`, {
    method: 'POST',
    // Do not set Content-Type header manually when using FormData, browser will set it with boundary
    body: formData,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
  }

  return response.json();
}
