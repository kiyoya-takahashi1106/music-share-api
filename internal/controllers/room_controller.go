package controllers

import (
	"net/http"

	"music-share-api/internal/repositories"
	"music-share-api/internal/services"

	"github.com/gin-gonic/gin"
)

type RoomController struct {
	roomService services.RoomService
}

func NewRoomController(roomService services.RoomService) *RoomController {
	return &RoomController{
		roomService: roomService,
	}
}

type CreateRoomRequest struct {
	RoomName            string  `json:"room_name" binding:"required"`
	IsPublic            bool    `json:"is_public" binding:"required"`
	RoomPassword        *string `json:"room_password"` // null許容
	Title               string  `json:"title"`
	Genre               string  `json:"genre"`
	MaxParticipants     int     `json:"max_participants" binding:"required"`
	HostUserID          int     `json:"host_user_id" binding:"required"`
	HostUserName        string  `json:"host_user_name" binding:"required"`
	PlayingPlaylistName string  `json:"playing_playlist_name"`
	PlayingSongName     string  `json:"playing_song_name"`
}

// room作ってmysqlに保存する
func (ctrl *RoomController) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input",
		})
		return
	}

	input := repositories.RoomCreateInput{
		RoomName:            req.RoomName,
		IsPublic:            req.IsPublic,
		RoomPassword:        req.RoomPassword,
		Title:               req.Title,
		Genre:               req.Genre,
		MaxParticipants:     req.MaxParticipants,
		HostUserID:          req.HostUserID,
		HostUserName:        req.HostUserName,
		PlayingPlaylistName: req.PlayingPlaylistName,
		PlayingSongName:     req.PlayingSongName,
	}

	roomID, err := ctrl.roomService.CreateRoom(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create room",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Room successfully created",
		"room_id": roomID,
	})
}
