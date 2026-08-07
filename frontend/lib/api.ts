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

export async function analyzeText(text: string): Promise<ScamAnalysisResponse> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  
  const response = await fetch(`${apiUrl}/api/v1/analyze`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ text_prompt: text }),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
  }

  return response.json();
}
