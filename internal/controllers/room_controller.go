package controllers

import (
	"log"
	"net/http"
	"strconv"

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

// CreateRoomRequest を repositories.RoomCreateInput のエイリアスにする
type CreateRoomRequest = repositories.RoomCreateInput

func (ctrl *RoomController) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input",
		})
		return
	}
	

	roomID, err := ctrl.roomService.CreateRoom(req)
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
		"roomId":  roomID,
	})
}


// JoinRoomRequest は /room/join エンドポイントのリクエストを表します
type JoinRoomRequest struct {
	UserID       int     `json:"userId" binding:"required"`
	UserName     string  `json:"userName" binding:"required"`
	RoomID       int     `json:"roomId" binding:"required"`
	RoomPassword *string `json:"roomPassword"`
}

// JoinRoom は /room/join エンドポイントを処理します
func (ctrl *RoomController) JoinRoom(c *gin.Context) {
	var req JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input",
		})
		return
	}

	// サービスを呼び出してルームへの参加を処理
	err := ctrl.roomService.JoinRoom(req.UserID, req.UserName, req.RoomID, req.RoomPassword)
	if err != nil {
		log.Printf("Failed to join room: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(), // サービスから返されたエラーメッセージをそのまま返す
		})
		return	
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Join",
	})
}


type LeaveRoomRequest struct {
	UserID int `json:"userId" binding:"required"`
	RoomID int `json:"roomId" binding:"required"`
}

func (ctrl *RoomController) LeaveRoom(c *gin.Context) {
	var req LeaveRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input",
		})
		return
	}

	// サービスを呼び出してルームからの退出を処理
	_, err := ctrl.roomService.LeaveRoom(req.UserID, req.RoomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Leave",
	})
}


// DeleteRoom はDELETEメソッドでルームを削除します。
// URL例: DELETE /room/1
func (ctrl *RoomController) DeleteRoom(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid roomId",
		})
		return
	}

	if err := ctrl.roomService.DeleteRoom(roomID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Delete",
	})
}


func (ctrl *RoomController) GetRoom(c *gin.Context) {
	roomIDStr := c.Param("roomId")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid roomId",
		})
		return
	}

	room, err := ctrl.roomService.GetRoom(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// 詳細な部屋情報を返す（Create/Joinと同じJSON形式）
	c.JSON(http.StatusOK, gin.H{
		"status":              "success",
		"message":             "Room successfully retrieved",
		"roomId":              room.RoomID,
		"roomName":            room.RoomName,
		"isPublic":            room.IsPublic,
		"genre":               room.Genre,
		"maxParticipants":     room.MaxParticipants,
		"nowParticipants":     room.NowParticipants,
		"host":                gin.H{"hostId": room.HostUserID, "hostName": room.HostUserName},
		"playingPlaylistName": room.PlayingPlaylistName,
		"playingSongName":     room.PlayingSongName,
		// Redis側の情報をフラットに展開
        "roomStatus":        room.RedisData.RoomStatus,
        "playingPlaylistId": room.RedisData.PlaylistID,
        "playingSongId":     room.RedisData.SongID,
        "updateSongAt":      room.RedisData.UpdateSongAt,
        "participants":      room.RedisData.Participants,
	})
}
