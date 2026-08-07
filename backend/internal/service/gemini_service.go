package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"shieldvn-backend/internal/model"

	"google.golang.org/genai"
)

type GeminiService struct {
	client *genai.Client
}

func NewGeminiService(ctx context.Context, apiKey string) (*GeminiService, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}
	return &GeminiService{client: client}, nil
}

func (s *GeminiService) Analyze(ctx context.Context, prompt string, imageBytes []byte, mimeType string) (*model.AnalysisResult, error) {
	modelName := "gemini-2.5-flash"

	systemInstruction := `You are ShieldVN, a privacy-first scam detection assistant for Vietnamese users. 
Analyze the provided text and/or image for potential scams, focusing on common typologies in Vietnam (e.g., CTV recruitment scams requiring deposits, fake bills/transfers, impersonation of officials or VNeID). 
If an image is provided (e.g. a transfer bill), check for signs of forgery like mismatched fonts, missing FT codes, or incorrect layouts.
Extract any entities like bank accounts, phone numbers, or URLs.
Return the result strictly as a structured JSON object matching the provided schema.`

	// Define the schema for structured output
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"risk_score": {
				Type:        genai.TypeString,
				Description: "One of: GREEN, YELLOW, RED",
			},
			"confidence_score": {
				Type:        genai.TypeNumber,
				Description: "Confidence score between 0.0 and 1.0",
			},
			"detected_patterns": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"evidence": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"recommendations": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"extracted_entities": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"bank_account": {Type: genai.TypeString},
					"phone_number": {Type: genai.TypeString},
					"url":          {Type: genai.TypeString},
				},
			},
		},
		Required: []string{"risk_score", "confidence_score", "detected_patterns", "evidence", "recommendations", "extracted_entities"},
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: systemInstruction},
			},
		},
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	var parts []*genai.Part
	if prompt != "" {
		parts = append(parts, &genai.Part{Text: prompt})
	}
	if len(imageBytes) > 0 {
		parts = append(parts, genai.NewPartFromBytes(imageBytes, mimeType))
	}

	// Retry logic (1 retry fallback)
	var lastErr error
	for i := 0; i < 2; i++ {
		resp, err := s.client.Models.GenerateContent(ctx, modelName, []*genai.Content{
			{
				Parts: parts,
			},
		}, config)
		if err != nil {
			lastErr = fmt.Errorf("GenerateContent API call failed: %w", err)
			continue
		}

		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			lastErr = fmt.Errorf("empty response from model")
			continue
		}

		jsonText := resp.Candidates[0].Content.Parts[0].Text
		if jsonText == "" {
			lastErr = fmt.Errorf("response text is empty")
			continue
		}

		var result model.AnalysisResult
		if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
			slog.Warn("failed to unmarshal Gemini JSON response, retrying", "attempt", i+1, "error", err, "json", jsonText)
			lastErr = fmt.Errorf("json unmarshal failed: %w", err)
			continue
		}

		// Success
		return &result, nil
	}

	return nil, fmt.Errorf("analysis failed after retries: %w", lastErr)
}
