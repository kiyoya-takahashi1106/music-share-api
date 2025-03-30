package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type ServiceRepository interface {
	InsertUserService(userID int, serviceName, serviceUserID, encryptedAccessToken, encryptedRefreshToken string, expiresAt time.Time) error
}

type serviceRepository struct {
	DB *sql.DB
}

func NewServiceRepository(db *sql.DB) ServiceRepository {
	return &serviceRepository{DB: db}
}

func (r *serviceRepository) InsertUserService(userID int, serviceName, serviceUserID, encryptedAccessToken, encryptedRefreshToken string, expiresAt time.Time) error {
    insertQuery := `
        INSERT INTO trx_users_services
        (user_id, service_name, service_user_id, encrypted_access_token, encrypted_refresh_token, expires_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
    _, err := r.DB.Exec(insertQuery, userID, serviceName, serviceUserID, encryptedAccessToken, encryptedRefreshToken, expiresAt)
    if err != nil {
        return fmt.Errorf("failed to insert user service: %w", err)
    }

    log.Printf("kiyoyakiyoya")
    log.Printf("kiyoyakiyoya")
    log.Printf("kiyoyakiyoya")

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