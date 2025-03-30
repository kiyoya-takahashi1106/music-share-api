package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// UserServiceData は１サービスの連携情報を表します。
type UserServiceData struct {
	ServiceUserID        string    `json:"serviceUserId"`
	EncryptedAccessToken string    `json:"encryptedAccessToken"`
	ExpiresAt            time.Time `json:"expiresAt"`
}

type AuthRepository interface {
	// GetUserInfo は、ユーザー基本情報と連携サービス情報を返します。
	GetUserInfo(userID int) (string, string, string, bool, map[string]UserServiceData, error)
	CreateUser(userName, email, hashedPassword string) (int, error)
	GetUserByEmail(email string) (int, string, string, string, bool, error)
	UpdateUserProfile(userID int, userName, email string) error
}

type authRepository struct {
	DB *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{DB: db}
}

// userIDからユーザー基本情報と連携サービス情報を取得
func (r *authRepository) GetUserInfo(userID int) (string, string, string, bool, map[string]UserServiceData, error) {
	var userName, email, role string
	var isSpotify bool

	query := `
        SELECT user_name, email, role, is_spotify
        FROM trx_users
        WHERE user_id = ?
        LIMIT 1
    `
	err := r.DB.QueryRow(query, userID).Scan(&userName, &email, &role, &isSpotify)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found for userID: %d", userID)
			return "", "", "", false, nil, fmt.Errorf("user not found")
		}
		log.Printf("Error retrieving user info for userID %d: %v", userID, err)
		return "", "", "", false, nil, fmt.Errorf("error retrieving user info: %w", err)
	}

	// 連携サービス情報を取得（1ユーザーにつき各サービスは１件前提）
	serviceQuery := `
        SELECT service_name, service_user_id, encrypted_access_token, expires_at
        FROM trx_users_services
        WHERE user_id = ? AND deleted_at IS NULL
    `
	rows, err := r.DB.Query(serviceQuery, userID)
	if err != nil {
		return "", "", "", false, nil, fmt.Errorf("failed to query user services: %w", err)
	}
	defer rows.Close()

	services := make(map[string]UserServiceData)
	for rows.Next() {
		var serviceName string
		var data UserServiceData
		if err := rows.Scan(&serviceName, &data.ServiceUserID, &data.EncryptedAccessToken, &data.ExpiresAt); err != nil {
			return "", "", "", false, nil, fmt.Errorf("failed to scan service row: %w", err)
		}
		services[serviceName] = data
	}
	return userName, email, role, isSpotify, services, nil
}

func (r *authRepository) CreateUser(userName, email, hashedPassword string) (int, error) {
	// 既存のユーザーが存在するか確認
	_, _, _, _, _, err := r.GetUserByEmail(email)
	if err == nil {
		// ユーザーが見つかった場合は既に登録されているのでエラーを返す
		return 0, fmt.Errorf("user with email %s already exists", email)
	}

	query := `
        INSERT INTO trx_users 
        (user_name, email, hash_password, profile_image_url, role, is_spotify)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	// profile_image_url: 空文字、role: "user", is_verified: false を設定
	result, err := r.DB.Exec(query, userName, email, hashedPassword, "", "user", false)
	if err != nil {
		log.Println("Error inserting user:", err)
		return 0, fmt.Errorf("error creating user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert ID: %v", err)
	}

	return int(userID), nil
}

func (r *authRepository) GetUserByEmail(email string) (int, string, string, string, bool, error) {
	var userID int
	var userName, hashedPassword, role string
	var isSpotify bool

	query := `
        SELECT user_id, user_name, hash_password, role, is_spotify
        FROM trx_users 
        WHERE email = ?
        LIMIT 1
    `
	err := r.DB.QueryRow(query, email).Scan(&userID, &userName, &hashedPassword, &role, &isSpotify)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", "", "", false, fmt.Errorf("user not found")
		}
		return 0, "", "", "", false, fmt.Errorf("error retrieving user: %v", err)
	}

	return userID, userName, hashedPassword, role, isSpotify, nil
}

func (r *authRepository) UpdateUserProfile(userID int, userName, email string) error {
	query := `
        UPDATE trx_users
        SET user_name = ?, email = ?
        WHERE user_id = ?
    `
	_, err := r.DB.Exec(query, userName, email, userID)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %v", err)
	}
	return nil
}
