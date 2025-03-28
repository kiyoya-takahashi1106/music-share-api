package repositories

import (
    "database/sql"
    "fmt"
    "log"
    "time"
)

// Room はルーム情報を表します
type Room struct {
    RoomID              int          `db:"room_id" json:"roomId"`
    RoomName            string       `db:"room_name" json:"roomName"`
    IsPublic            bool         `db:"is_public" json:"isPublic"`
    Genre               string       `db:"genre" json:"genre"`
    PlayingPlaylistName string       `db:"playing_playlist_name" json:"playingPlaylistName"`
    PlayingSongName     string       `db:"playing_song_name" json:"playingSongName"`
    MaxParticipants     int          `db:"max_participants" json:"maxParticipants"`
    NowParticipants     int          `db:"now_participants" json:"nowParticipants"`
    HostUserID          int          `db:"host_user_id" json:"hostUserId"`
    HostUserName        string       `db:"host_user_name" json:"hostUserName"`
    CreateAt            time.Time    `db:"created_at" json:"createAt"`
    UpdateAt            time.Time    `db:"updated_at" json:"updateAt"`
    DeletedAt           sql.NullTime `db:"deleted_at" json:"deletedAt"`
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