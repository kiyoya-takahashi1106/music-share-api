package controllers

import (
	"net/http"
	"music-share-api/internal/services"

	"github.com/gin-gonic/gin"
)

type ServiceController struct {
	spotifyService services.SpotifyService
}

func NewServiceController(spotifyService services.SpotifyService) *ServiceController {
	return &ServiceController{spotifyService: spotifyService}
}


// SpotifyConnct 
func (ctrl *ServiceController) SpotifyConnct(c *gin.Context) {
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
