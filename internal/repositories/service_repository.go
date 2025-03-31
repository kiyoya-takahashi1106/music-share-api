package repositories

import (
	"database/sql"
	"fmt"
	"time"
)

type ServiceRepository interface {
	InsertUserService(userID int, serviceName, serviceUserID, serviceUserName, encryptedAccessToken, encryptedRefreshToken string, expiresAt time.Time) error
	DeleteUserService(userID int, serviceName string) error
	// 新規追加：Spotifyのリフレッシュトークンを取得する
	GetSpotifyRefreshToken(userID int) (string, error)
	// 新規追加：Spotifyのアクセストークン・リフレッシュトークン・有効期限を更新する
	UpdateSpotifyToken(userID int, newAccessToken, newRefreshToken string, newExpiresAt time.Time) error
}

type serviceRepository struct {
	DB *sql.DB
}

func NewServiceRepository(db *sql.DB) ServiceRepository {
	return &serviceRepository{DB: db}
}

func (r *serviceRepository) InsertUserService(userID int, serviceName, serviceUserID, serviceUserName, encryptedAccessToken, encryptedRefreshToken string, expiresAt time.Time) error {
	// JSTの現在時刻を取得（created_at, updated_atとして使う）
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return fmt.Errorf("failed to load JST location: %w", err)
	}
	currentTime := time.Now().In(loc)

	// この例では、created_atとupdated_atをcurrentTimeとして明示的に設定する
	query := `
        INSERT INTO trx_users_services
        (user_id, service_name, service_user_id, service_user_name, encrypted_access_token, encrypted_refresh_token, expires_at, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	// expiresAt, currentTimeのフォーマットは "2006-01-02 15:04:05" を使用
	formattedExpiresAt := expiresAt.Format("2006-01-02 15:04:05")
	formattedCurrentTime := currentTime.Format("2006-01-02 15:04:05")

	_, err = r.DB.Exec(query, userID, serviceName, serviceUserID, serviceUserName, encryptedAccessToken, encryptedRefreshToken, formattedExpiresAt, formattedCurrentTime, formattedCurrentTime)
	if err != nil {
		return fmt.Errorf("failed to insert user service: %w", err)
	}

	// trx_users の is_spotify を更新（必要な場合）
	updateQuery := `
        UPDATE trx_users 
        SET is_spotify = ? 
        WHERE user_id = ?
    `
	_, err = r.DB.Exec(updateQuery, true, userID)
	if err != nil {
		return fmt.Errorf("failed to update is_spotify status: %w", err)
	}
	return nil
}

func (r *serviceRepository) DeleteUserService(userID int, serviceName string) error {
	deleteQuery := `
        DELETE FROM trx_users_services
        WHERE user_id = ? AND service_name = ?
    `
	_, err := r.DB.Exec(deleteQuery, userID, serviceName)
	if err != nil {
		return fmt.Errorf("failed to delete user service: %w", err)
	}

	updateQuery := `
        UPDATE trx_users 
        SET is_spotify = ? 
        WHERE user_id = ?
    `
	_, err = r.DB.Exec(updateQuery, false, userID)
	if err != nil {
		return fmt.Errorf("failed to update is_spotify: %w", err)
	}
	return nil
}

// GetSpotifyRefreshToken userIdから refresh token を取得します。
func (r *serviceRepository) GetSpotifyRefreshToken(userID int) (string, error) {
	query := `
        SELECT encrypted_refresh_token
        FROM trx_users_services
        WHERE user_id = ? AND service_name = ? AND deleted_at IS NULL
        LIMIT 1
    `
	var encryptedRefreshToken string
	err := r.DB.QueryRow(query, userID, "spotify").Scan(&encryptedRefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}
	return encryptedRefreshToken, nil
}

// UpdateSpotifyToken は、trx_users_services内のアクセストークン、リフレッシュトークン、有効期限を更新します。
func (r *serviceRepository) UpdateSpotifyToken(userID int, newAccessToken, newRefreshToken string, newExpiresAt time.Time) error {
	// JSTの日時文字列（例："2006-01-02 15:04:05"）
	formattedExpiresAt := newExpiresAt.Format("2006-01-02 15:04:05")

	query := `
         UPDATE trx_users_services
         SET encrypted_access_token = ?,
             encrypted_refresh_token = ?,
             expires_at = ?
         WHERE user_id = ? AND service_name = ?
    `
	_, err := r.DB.Exec(query, newAccessToken, newRefreshToken, formattedExpiresAt, userID, "spotify")
	if err != nil {
		return fmt.Errorf("failed to update spotify token: %w", err)
	}
	return nil
}
