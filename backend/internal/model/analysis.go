package model

// AnalysisInput represents the incoming payload from the frontend for text analysis.
type AnalysisInput struct {
	TextPrompt string `json:"text_prompt" binding:"required"`
}

// ExtractedEntities holds any structured data found in the text (like bank accounts, phone numbers, URLs).
type ExtractedEntities struct {
	BankAccount string `json:"bank_account,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	URL         string `json:"url,omitempty"`
}

// AnalysisResult represents the structured JSON output returned by the Gemini AI analysis.
type AnalysisResult struct {
	RiskScore        string            `json:"risk_score"`
	ConfidenceScore  float64           `json:"confidence_score"`
	DetectedPatterns []string          `json:"detected_patterns"`
	Evidence         []string          `json:"evidence"`
	Recommendations  []string          `json:"recommendations"`
	ExtractedEntities ExtractedEntities `json:"extracted_entities"`
}
