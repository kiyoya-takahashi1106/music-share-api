package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RoomCreateInput struct {
	RoomName            string          `json:"roomName"`
	IsPublic            bool            `json:"isPublic"`
	RoomPassword        *string         `json:"roomPassword"`
	Genre               string          `json:"genre"`
	MaxParticipants     int             `json:"maxParticipants"`
	HostUserID          int             `json:"hostUserId"`
	HostUserName        string          `json:"hostUserName"`
	PlayingPlaylistID   string          `json:"playingPlaylistId"`
	PlayingPlaylistName string          `json:"playingPlaylistName"`
	PlayingSongID       string          `json:"playingSongId"`
	PlayingSongName     string          `json:"playingSongName"`
	Songs               map[string]Song `json:"songs"`
}

type Song struct {
	SongId       string `json:"songId"`
	SongName     string `json:"songName"`
	Artist       string `json:"artist"`
	SongLength   int    `json:"songLength"`
	SongImageUrl string `json:"songImageUrl"`
}

// RedisRoomParticipant はRedis内の参加者情報を表します
type RedisRoomParticipant struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
}

// RedisRoomData はRedisに保存するルーム情報を表します
type RedisRoomData struct {
	RoomStatus   string                 `json:"roomStatus"`
	PlaylistID   string                 `json:"playingPlaylistId"`
	SongID       string                 `json:"playingSongId"`
	UpdateSongAt string                 `json:"updateSongAt"`
	Participants []RedisRoomParticipant `json:"participants"`
}

// RoomAllInfo はルーム情報を表します（MySQLとRedisのデータを統合）
type RoomAllInfo struct {
	RoomID              int           `db:"room_id" json:"roomId"`
	RoomName            string        `db:"room_name" json:"roomName"`
	IsPublic            bool          `db:"is_public" json:"isPublic"`
	Genre               string        `db:"genre" json:"genre"`
	PlayingPlaylistName string        `db:"playing_playlist_name" json:"playingPlaylistName"`
	PlayingSongName     string        `db:"playing_song_name" json:"playingSongName"`
	MaxParticipants     int           `db:"max_participants" json:"maxParticipants"`
	NowParticipants     int           `db:"now_participants" json:"nowParticipants"`
	HostUserID          int           `db:"host_user_id" json:"hostUserId"`
	HostUserName        string        `db:"host_user_name" json:"hostUserName"`
	CreateAt            time.Time     `db:"created_at" json:"createAt"`
	UpdateAt            time.Time     `db:"updated_at" json:"updateAt"`
	DeletedAt           sql.NullTime  `db:"deleted_at" json:"deletedAt"`
	RedisData           RedisRoomData `json:"redisData"` // Redisからの情報
}

type RoomRepository interface {
	CreateRoom(input RoomCreateInput) (int, error)
	JoinRoom(userID int, userName string, roomID int, roomPassword *string) error
	LeaveRoom(userID int, roomID int) (*RoomAllInfo, error)
	DeleteRoom(roomID int) error
	GetRoomByID(roomID int) (*RoomAllInfo, error)
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

func (r *roomRepository) CreateRoom(input RoomCreateInput) (int, error) {
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
		return 0, fmt.Errorf("failed to insert room: %v", err)
	}

	roomID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get room id: %v", err)
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
		return int(roomID), fmt.Errorf("failed to marshal Redis data: %v", err)
	}

	err = r.RedisClient.Set(ctx, key, redisJSON, 0).Err() // 有効期限なし
	if err != nil {
		return int(roomID), fmt.Errorf("failed to save to Redis: %v", err)
	}

	fmt.Println("fgeshrdhrrethr")

	// trx_rooms_songs に songs を挿入
	for idx, song := range input.Songs {
		insertSongQuery := `
            INSERT INTO trx_rooms_songs
            (room_id, song_index, song_id, song_name, artist, song_length, song_image_url)
            VALUES (?, ?, ?, ?, ?, ?, ?)
        `
		if _, err := r.DB.Exec(insertSongQuery,
			roomID,
			idx, // songs のキー (何番目か)
			song.SongId,
			song.SongName,
			song.Artist,
			song.SongLength, // DB で timestamp 型の場合は必要に応じて変換
			song.SongImageUrl,
		); err != nil {
			fmt.Println(err)
			return int(roomID), fmt.Errorf("failed to insert room song: %v", err)
		}
	}

	return int(roomID), nil
}

// JoinRoom は、ユーザーをルームに参加させるロジックを実装します。
func (r *roomRepository) JoinRoom(userID int, userName string, roomID int, roomPassword *string) error {
	// 1. MySQLからルームの基本情報を取得
	var room RoomAllInfo
	query := `
        SELECT room_id, room_name, is_public, genre, max_participants, now_participants, 
               host_user_id, host_user_name, playing_playlist_name, playing_song_name 
        FROM trx_rooms 
        WHERE room_id = ?`
	err := r.DB.QueryRow(query, roomID).Scan(
		&room.RoomID, &room.RoomName, &room.IsPublic, &room.Genre, &room.MaxParticipants, &room.NowParticipants,
		&room.HostUserID, &room.HostUserName, &room.PlayingPlaylistName, &room.PlayingSongName,
	)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}

	// 2. 参加人数が上限に達していないか確認
	if room.NowParticipants >= room.MaxParticipants {
		return fmt.Errorf("room is full")
	}

	// 3. Redisからルームデータ（参加者リストなど）を取得
	ctx := context.Background()
	key := fmt.Sprintf("room:%d", roomID)
	val, err := r.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get room data from Redis: %w", err)
	}

	var redisData RedisRoomData
	if err = json.Unmarshal([]byte(val), &redisData); err != nil {
		return fmt.Errorf("failed to unmarshal Redis data: %w", err)
	}

	// 4. 新しい参加者を追加（ユーザーIDは文字列に変換）
	newParticipant := RedisRoomParticipant{
		UserID:   fmt.Sprintf("%d", userID),
		Username: userName,
	}
	redisData.Participants = append(redisData.Participants, newParticipant)

	// 5. 参加者追加後、Redisに更新されたデータを保存
	updatedRedisJSON, err := json.Marshal(redisData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated Redis data: %w", err)
	}
	if err = r.RedisClient.Set(ctx, key, updatedRedisJSON, 0).Err(); err != nil {
		return fmt.Errorf("failed to update room data in Redis: %w", err)
	}

	// 6. MySQLの参加者数(now_participants)を更新
	updateQuery := `UPDATE trx_rooms SET now_participants = now_participants + 1 WHERE room_id = ?`
	if _, err = r.DB.Exec(updateQuery, roomID); err != nil {
		return fmt.Errorf("failed to update participants in MySQL: %w", err)
	}

	return nil
}

// LeaveRoom は、ユーザーをルームから退出させるロジックを実装します。
func (r *roomRepository) LeaveRoom(userID int, roomID int) (*RoomAllInfo, error) {
	var room RoomAllInfo

	// 2. Redisからルームデータ（参加者リストなど）を取得
	ctx := context.Background()
	key := fmt.Sprintf("room:%d", roomID)
	val, err := r.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get room data from Redis: %w", err)
	}
	var redisData RedisRoomData
	if err = json.Unmarshal([]byte(val), &redisData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Redis data: %w", err)
	}

	// 3. 自分の参加者データを削除（ユーザーIDは文字列に変換）
	userIDStr := fmt.Sprintf("%d", userID)
	newParticipants := make([]RedisRoomParticipant, 0)
	for _, participant := range redisData.Participants {
		if participant.UserID != userIDStr {
			newParticipants = append(newParticipants, participant)
		}
	}
	redisData.Participants = newParticipants

	// 4. Redisに更新されたデータを保存
	updatedRedisJSON, err := json.Marshal(redisData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated Redis data: %w", err)
	}
	if err = r.RedisClient.Set(ctx, key, updatedRedisJSON, 0).Err(); err != nil {
		return nil, fmt.Errorf("failed to update room data in Redis: %w", err)
	}

	// 5. MySQLの参加者数(now_participants)を -1 し、deleted_at に現在のタイムスタンプを記録する
	updateQuery := `UPDATE trx_rooms SET now_participants = now_participants - 1 WHERE room_id = ?`
	if _, err = r.DB.Exec(updateQuery, roomID); err != nil {
		return nil, fmt.Errorf("failed to update participants in MySQL: %w", err)
	}

	return &room, nil
}

func (r *roomRepository) DeleteRoom(roomID int) error {
	now := time.Now()
	updateQuery := `UPDATE trx_rooms SET deleted_at = ? WHERE room_id = ?`
	_, err := r.DB.Exec(updateQuery, now, roomID)
	if err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}
	return nil
}

// GetRoomByID はroomIDからMySQLとRedisの情報を統合して部屋の詳細情報を取得します。
func (r *roomRepository) GetRoomByID(roomID int) (*RoomAllInfo, error) {
	var room RoomAllInfo

	// MySQLから部屋の詳細情報を取得
	query := `
        SELECT room_id, room_name, is_public, genre, playing_playlist_name, playing_song_name,
               max_participants, now_participants, host_user_id, host_user_name, created_at
        FROM trx_rooms 
        WHERE room_id = ?`
	err := r.DB.QueryRow(query, roomID).Scan(
		&room.RoomID, &room.RoomName, &room.IsPublic, &room.Genre,
		&room.PlayingPlaylistName, &room.PlayingSongName, &room.MaxParticipants,
		&room.NowParticipants, &room.HostUserID, &room.HostUserName, &room.CreateAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get room from MySQL: %w", err)
	}

	// Redisから部屋データ（参加者リストなど）を取得
	ctx := context.Background()
	key := fmt.Sprintf("room:%d", roomID)
	val, err := r.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get room data from Redis: %w", err)
	}
	var redisData RedisRoomData
	if err = json.Unmarshal([]byte(val), &redisData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Redis data: %w", err)
	}

	room.RedisData = redisData
	return &room, nil
}
