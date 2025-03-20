package repositories

import (
	"database/sql"
	"fmt"
	"time"
	"log"
)

// Room はルーム情報を表します
type Room struct {
	RoomID              int          `db:"room_id" json:"room_id"`
	RoomName            string       `db:"room_name" json:"room_name"`
	IsPublic            bool         `db:"is_public" json:"is_public"`
	Genre               string       `db:"genre" json:"genre"`
	PlayingPlaylistName string       `db:"playing_playlist_name" json:"playing_playlist_name"`
	PlayingSongName     string       `db:"playing_song_name" json:"playing_song_name"`
	MaxParticipants     int          `db:"max_participants" json:"max_participants"`
	NowParticipants     int          `db:"now_participants" json:"now_participants"`
	HostUserID          int          `db:"host_user_id" json:"host_user_id"`
    HostUserName        string       `db:"host_user_name" json:"host_user_name"`
	CreateAt            time.Time    `db:"created_at" json:"create_at"`
	UpdateAt            time.Time    `db:"updated_at" json:"update_at"`
	DeletedAt           sql.NullTime `db:"deleted_at" json:"deleted_at"`
}

type RoomsRepository interface {
	GetPublicRooms() ([]Room, error)
}

type roomsRepository struct {
	DB *sql.DB
}

func NewRoomsRepository(db *sql.DB) RoomsRepository {
	return &roomsRepository{DB: db}
}

func (r *roomsRepository) GetPublicRooms() ([]Room, error) {
    query := `
        SELECT room_id, room_name, is_public, genre, playing_playlist_name, playing_song_name,
               max_participants, now_participants, host_user_id, host_user_name, created_at, updated_at, deleted_at
        FROM trx_rooms
        WHERE is_public = ? AND deleted_at IS NULL
    `
    rows, err := r.DB.Query(query, true)
    if err != nil {
        log.Printf("DB query error: %v", err)
        return nil, fmt.Errorf("query error: %v", err)
    }
    defer rows.Close()

    var rooms []Room
    for rows.Next() {
        var room Room
        if err := rows.Scan(
            &room.RoomID,
            &room.RoomName,
            &room.IsPublic,
            &room.Genre,
            &room.PlayingPlaylistName,
            &room.PlayingSongName,
            &room.MaxParticipants,
            &room.NowParticipants,
            &room.HostUserID,
            &room.HostUserName,
            &room.CreateAt,
            &room.UpdateAt,
            &room.DeletedAt,
        ); err != nil {
            log.Printf("Row scan error: %v", err)
            return nil, fmt.Errorf("scan error: %v", err)
        }
        rooms = append(rooms, room)
    }

    if err := rows.Err(); err != nil {
        log.Printf("Rows error: %v", err)
        return nil, fmt.Errorf("rows error: %v", err)
    }

    return rooms, nil
}