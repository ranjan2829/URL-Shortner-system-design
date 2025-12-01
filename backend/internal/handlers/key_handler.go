package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/services"
)

type KeyHandler struct {
	keyService *services.KeyService
}

func NewKeyHandler(keyService *services.KeyService) *KeyHandler {
	return &KeyHandler{
		keyService: keyService,
	}
}

type GenerateResponse struct {
	ShortCode string `json:"short_code"`
}

// GenerateKey handles GET /api/v1/generate
// This generates a new short code
func (h *KeyHandler) GenerateKey(c *gin.Context) {
	shortCode := h.keyService.GenerateShortCode()
	
	response := GenerateResponse{
		ShortCode: shortCode,
	}
	
	c.JSON(http.StatusOK, response)
}

