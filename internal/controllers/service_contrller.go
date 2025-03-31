package controllers

import (
	"log"
	"music-share-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceController struct {
	spotifyService services.SpotifyService
}

func NewServiceController(spotifyService services.SpotifyService) *ServiceController {
	return &ServiceController{spotifyService: spotifyService}
}

// SpotifyConnct
func (ctrl *ServiceController) SpotifyConnect(c *gin.Context) {
	var req struct {
		UserID int    `json:"userId" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid input"})
		return
	}

	if err := ctrl.spotifyService.ConnectSpotify(req.UserID, req.Code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Spotify connected successfully",
	})
}

func (ctrl *ServiceController) DisconnectSpotify(c *gin.Context) {
	log.Println("DisconnectSpotify called")
	var req struct {
		UserID int `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid input"})
		return
	}

	if err := ctrl.spotifyService.DeleteSpotify(req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Account deleted successfully",
	})
}

// RefreshSpotifyToken
func (ctrl *ServiceController) RefreshSpotifyToken(c *gin.Context) {
	var req struct {
		UserID int `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid input"})
		return
	}

	encryptedAccessToken, newExpiresAt, err := ctrl.spotifyService.RefreshSpotifyToken(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// ここでは Format("2006-01-02T15:04:05Z") でISO8601風に変換
	formattedExpiresAt := newExpiresAt.Format("2006-01-02T15:04:05Z")

	c.JSON(http.StatusOK, gin.H{
		"status":               "success",
		"message":              "Token refresh successfully",
		"encryptedAccessToken": encryptedAccessToken,
		"expiresAt":            formattedExpiresAt,
	})
}
