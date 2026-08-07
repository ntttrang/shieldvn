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

	console.log("🚀 [Frontend] Sending request to Gemini via Backend...", { 
		textLength: text.length, 
		hasImage: !!image,
		imageSize: image ? image.size : 0 
	});

	const response = await fetch(`${apiUrl}/api/v1/analyze`, {
    method: 'POST',
    // Do not set Content-Type header manually when using FormData, browser will set it with boundary
    body: formData,
  });

	if (!response.ok) {
		const errorData = await response.json().catch(() => ({}));
		console.error("❌ [Frontend] Backend error:", errorData);
		throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
	}

	const result = await response.json();
	console.log("✅ [Frontend] Received analysis result:", result);
	return result;
}
