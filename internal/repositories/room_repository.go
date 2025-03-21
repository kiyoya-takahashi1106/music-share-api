package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RoomCreateInput はルーム作成時の入力データを表します
type RoomCreateInput struct {
	RoomName            string
	IsPublic            bool
	RoomPassword        *string // null許容
	Title               string
	Genre               string
	MaxParticipants     int
	HostUserID          int
	HostUserName        string
	PlayingPlaylistName string
	PlayingPlaylistID   string // Redis保存用に追加
	PlayingSongName     string
	PlayingSongID       string // Redis保存用に追加
}

// RedisRoomParticipant はRedis内の参加者情報を表します
type RedisRoomParticipant struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// RedisRoomData はRedisに保存するルーム情報を表します
type RedisRoomData struct {
	RoomStatus   string                 `json:"room_status"`
	PlaylistID   string                 `json:"playing_playlist_id"`
	SongID       string                 `json:"playing_song_id"`
	UpdateSongAt string                 `json:"update_song_at"`
	Participants []RedisRoomParticipant `json:"participants"`
}

type RoomRepository interface {
	CreateRoom(input RoomCreateInput) (int, *RedisRoomData, error)
}

type roomRepository struct {
	DB          *sql.DB
	RedisClient *redis.Client
}

func NewRoomRepository(db *sql.DB, redisClient *redis.Client) RoomRepository {
	return &roomRepository{
		DB:          db,
		RedisClient: redisClient,
	}
}

func (r *roomRepository) CreateRoom(input RoomCreateInput) (int, *RedisRoomData, error) {
	// MySQL にルーム情報を保存
	query := `
        INSERT INTO trx_rooms 
        (room_name, is_public, room_password, genre, max_participants, now_participants, host_user_id, host_user_name, playing_playlist_name, playing_song_name)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	result, err := r.DB.Exec(query,
		input.RoomName,
		input.IsPublic,
		input.RoomPassword,
		input.Genre,
		input.MaxParticipants,
		1, // now_participants は 1（ルーム作成者）
		input.HostUserID,
		input.HostUserName,
		input.PlayingPlaylistName,
		input.PlayingSongName,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to insert room: %v", err)
	}

	roomID, err := result.LastInsertId()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get room id: %v", err)
	}

	// Redis用のデータを作成
	now := time.Now().Format("200601021504") // YYYYMMDDHHmm形式
	redisData := &RedisRoomData{
		RoomStatus:   "playing",
		PlaylistID:   input.PlayingPlaylistID,
		SongID:       input.PlayingSongID,
		UpdateSongAt: now,
		Participants: []RedisRoomParticipant{}, // 初期状態では参加者なし（ホストは別管理）
	}

	// Redisにデータを保存
	ctx := context.Background()
	key := fmt.Sprintf("room:%d", roomID)

	redisJSON, err := json.Marshal(redisData)
	if err != nil {
		return int(roomID), nil, fmt.Errorf("failed to marshal Redis data: %v", err)
	}

	err = r.RedisClient.Set(ctx, key, redisJSON, 0).Err() // 有効期限なし
	if err != nil {
		return int(roomID), nil, fmt.Errorf("failed to save to Redis: %v", err)
	}

	return int(roomID), redisData, nil
}
