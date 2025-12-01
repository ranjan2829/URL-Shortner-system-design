package handlers

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/ranjanshahajishitole/url-shortener/backend/internal/services"

)

type URLHandler struct {
	urlService *services.URLService
}

func NewURLHandler(urlService *services.URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

type ShortenURLRequest struct {
	URL 	string 	`json:"url" binding:"required,url"`
	ExpiresIn *int `json:"expires_in,omitempty"`

}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	ShortCode string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

func (h * URLHandler) ShortenURL(c *gin.Context) {
	var req ShortenURLRequest
	if err := c.ShouldBindJSON(&req);err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return
	}
	var expiresIN *time.Duration
	if req.ExpiresIn !=nil{
		duration := time.Duration(*req.ExpiresIn) * time.Hour
		expiresIN =&duration
	}
	shortURL,err:=h.urlService.ShortenURL(c.Request.Context(),req.URL,expiresIN)
	if err!=nil{
		if err ==services.ErrInvalidURL{
			c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid URL"})
			return
		}
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to shorten URL"})
		return
	}
	response:=ShortenResponse{
		ShortURL: fmt.Sprintf("http://localhost:8080/%s",shortURL.ShortCode),
		ShortCode: shortURL.ShortCode,
		OriginalURL: shortURL.OriginalURL,
		ExpiresAt: shortURL.ExpiresAt,
	}
}


func (h * URLHandler) RedirectURL(c * gin.Context) {
	shortCode := c.Param("code")
	if shortCode==""{
		c.JSON(http.StatusBadRequest,gin.H{"error":"short code in needed"})
		return
	}
	originalURL,err:=h.urlService.GetOriginalURL(c.Request.Context(),shortCode)
	if err!=nil{
		if err == services.ErrURLNotFound{
			c.JSON(http.StatusNotFound,gin.H{"error":"URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to redirect URL"})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect,originalURL)


}func (h *URLHandler) GetStats(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	stats, err := h.urlService.GetStats(c.Request.Context(), shortCode)
	if err != nil {
		if err == services.ErrURLNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}