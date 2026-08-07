package handler

import (
	"net/http"

	"shieldvn-backend/internal/model"
	"shieldvn-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AnalyzeHandler handles text analysis requests.
func AnalyzeHandler(geminiSvc *service.GeminiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input model.AnalysisInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		result, err := geminiSvc.Analyze(c.Request.Context(), input.TextPrompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze text", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
