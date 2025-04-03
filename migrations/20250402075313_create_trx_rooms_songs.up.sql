CREATE TABLE trx_rooms_songs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    room_id INT NOT NULL,
    song_index INT NOT NULL,
    song_id VARCHAR(255) NOT NULL,
    song_name VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    song_length TIMESTAMP NOT NULL,
    song_image_url VARCHAR(512),
    FOREIGN KEY (room_id) REFERENCES trx_rooms(room_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
