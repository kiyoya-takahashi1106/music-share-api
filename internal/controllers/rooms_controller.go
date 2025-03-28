package controllers

import (
	"net/http"

	"music-share-api/internal/services"

	"github.com/gin-gonic/gin"
)

type RoomsController struct {
	roomsService services.RoomsService
}

func NewRoomsController(roomsService services.RoomsService) *RoomsController {
	return &RoomsController{
		roomsService: roomsService,
	}
}

func (ctrl *RoomsController) GetPublicRooms(c *gin.Context) {
	rooms, err := ctrl.roomsService.GetPublicRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error fetching rooms",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Rooms matching search criteria",
		"rooms":   rooms,
	})
}
