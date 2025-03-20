package repositories

import (
	"database/sql"
	"fmt"
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
	PlayingSongName     string
}

type RoomRepository interface {
	CreateRoom(input RoomCreateInput) (int, error)
}

type roomRepository struct {
	DB *sql.DB
}

func NewRoomRepository(db *sql.DB) RoomRepository {
	return &roomRepository{DB: db}
}

func (r *roomRepository) CreateRoom(input RoomCreateInput) (int, error) {
	query := `
        INSERT INTO trx_rooms 
        (room_name, is_public, room_password, title, genre, max_participants, now_participants, host_user_id, host_user_name, playing_playlist_name, playing_song_name)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	// 新規作成時は参加者はホスト1名から開始
	result, err := r.DB.Exec(query,
		input.RoomName,
		input.IsPublic,
		input.RoomPassword,
		input.Title,
		input.Genre,
		input.MaxParticipants,
		1, // now_participants は 1（ルーム作成者）
		input.HostUserID,
		input.HostUserName,
		input.PlayingPlaylistName,
		input.PlayingSongName,
	)

	//　いつかtrx_joins Tableへのデータ挿入も...

	if err != nil {
		return 0, fmt.Errorf("failed to insert room: %v", err)
	}
	roomID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get room id: %v", err)
	}
	return int(roomID), nil
}
