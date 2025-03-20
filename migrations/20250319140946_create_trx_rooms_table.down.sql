CREATE TABLE trx_rooms (
    room_id INT AUTO_INCREMENT PRIMARY KEY,
    room_name VARCHAR(255) NOT NULL,
    is_public BOOLEAN NOT NULL,
    room_password VARCHAR(255),
    title VARCHAR(255),
    genre VARCHAR(255),
    max_participants INT NOT NULL DEFAULT 10,
    now_participants INT NOT NULL DEFAULT 1,
    host_user_id INT NOT NULL,
    host_user_name VARCHAR(255) NOT NULL,
    playing_playlist_name VARCHAR(255),
    playing_song_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (host_user_id) REFERENCES trx_users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
