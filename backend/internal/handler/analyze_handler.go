package handler

import (
	"io"
	"net/http"

	"shieldvn-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AnalyzeHandler handles multipart text and image analysis requests.
func AnalyzeHandler(geminiSvc *service.GeminiService, blacklistSvc *service.BlacklistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		textPrompt := c.PostForm("text_prompt")
		
		var imageBytes []byte
		var mimeType string

		fileHeader, err := c.FormFile("image")
		if err == nil && fileHeader != nil {
			// Check file size < 5MB
			if fileHeader.Size > 5*1024*1024 {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Image must be less than 5MB"})
				return
			}
			
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
				return
			}
			defer file.Close()
			
			bytes, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
				return
			}
			imageBytes = bytes
			
			contentType := fileHeader.Header.Get("Content-Type")
			if contentType != "" {
				mimeType = contentType
			} else {
				mimeType = http.DetectContentType(bytes)
			}
			
			if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported image format. Use JPEG, PNG, or WEBP."})
				return
			}
		}

		if textPrompt == "" && len(imageBytes) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Must provide text_prompt or image"})
			return
		}

		result, err := geminiSvc.Analyze(c.Request.Context(), textPrompt, imageBytes, mimeType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze input", "details": err.Error()})
			return
		}

		// Phase 03: Tier-1 Blacklist Lookup
		blacklistHits := blacklistSvc.CheckTier1(c.Request.Context(), result.ExtractedEntities)
		if len(blacklistHits) > 0 {
			result.Evidence = append(result.Evidence, blacklistHits...)
			// Escalate risk score to at least YELLOW if it's currently GREEN
			if result.RiskScore == "GREEN" {
				result.RiskScore = "YELLOW"
			}
			// Alternatively, you could escalate to RED, but we follow "at least YELLOW" as per plan
		}

		c.JSON(http.StatusOK, result)
	}
}
