package controllers

import (
	"log"
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
	Genre               string  `json:"genre"`
	MaxParticipants     int     `json:"max_participants" binding:"required"`
	HostUserID          int     `json:"host_user_id" binding:"required"`
	HostUserName        string  `json:"host_user_name" binding:"required"`
	PlayingPlaylistID   string     `json:"playing_playlist_id"` // 入力は受け取るがDBには保存しない例
	PlayingPlaylistName string  `json:"playing_playlist_name"`
	PlayingSongID       string     `json:"playing_song_id"` // 入力は受け取るがDBには保存しない例
	PlayingSongName     string  `json:"playing_song_name"`
}

// CreateRoom 新規ルーム作成 API
func (ctrl *RoomController) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input",
		})
		return
	}

	log.Printf("CreateRoomRequest: %+v", req)

	// DBへは playlist_id, song_id 以外の必要な項目を保存する
	input := repositories.RoomCreateInput{
		RoomName:            req.RoomName,
		IsPublic:            req.IsPublic,
		RoomPassword:        req.RoomPassword,
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

	// 出力例：　hostはネストしたJSONオブジェクトとして返す
	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"message":          "Room successfully created",
		"room_id":          roomID,
		"room_name":        req.RoomName,
		"is_public":        req.IsPublic,
		"genre":            req.Genre,
		"max_participants": req.MaxParticipants,
		"now_participants": 1, // 新規作成時は1（ルーム作成者）
		"host": gin.H{
			"host_id":   req.HostUserID,
			"host_name": req.HostUserName,
		},
	})
}
