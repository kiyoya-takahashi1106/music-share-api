package repositories

import (
	"database/sql"
	"fmt"
	"time"
)

type ServiceRepository interface {
	InsertUserService(userID int, serviceName, serviceUserID, serviceUserName, encryptedAccessToken, encryptedRefreshToken string, expiresAt time.Time) error
	DeleteUserService(userID int, serviceName string) error
}

type serviceRepository struct {
	DB *sql.DB
}

func NewServiceRepository(db *sql.DB) ServiceRepository {
	return &serviceRepository{DB: db}
}

func (r *serviceRepository) InsertUserService(userID int, serviceName, serviceUserID, serviceUserName, encryptedAccessToken, encryptedRefreshToken string, expiresAt time.Time) error {
	query := `
        INSERT INTO trx_users_services
        (user_id, service_name, service_user_id, service_user_name, encrypted_access_token, encrypted_refresh_token, expires_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	_, err := r.DB.Exec(query, userID, serviceName, serviceUserID, serviceUserName, encryptedAccessToken, encryptedRefreshToken, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert user service: %w", err)
	}

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
	// trx_users_services から対象レコードを削除
	deleteQuery := `
        DELETE FROM trx_users_services
        WHERE user_id = ? AND service_name = ?
    `
	_, err := r.DB.Exec(deleteQuery, userID, serviceName)
	if err != nil {
		return fmt.Errorf("failed to delete user service: %w", err)
	}

	// trx_users の is_spotify を false に更新
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
