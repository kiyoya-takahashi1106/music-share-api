package repositories

import (
	"database/sql"
	"fmt"
	"log"
)

type AuthRepository interface {
	GetUserInfo(userID int) (string, string, string, bool, error)
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

func CreateSetCookie(userId string) {

}

// userIDからuser情報取得
func (r *authRepository) GetUserInfo(userID int) (string, string, string, bool, error) {
    var userName, email, role string
    var isVerified bool

    query := `
        SELECT user_name, email, role, is_verified
        FROM trx_users
        WHERE user_id = ?
        LIMIT 1
    `
    err := r.DB.QueryRow(query, userID).Scan(&userName, &email, &role, &isVerified)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Printf("User not found for userID: %d", userID)
            return "", "", "", false, fmt.Errorf("user not found")
        }
        log.Printf("Error retrieving user info for userID %d: %v", userID, err)
        return "", "", "", false, fmt.Errorf("error retrieving user info: %w", err)
    }

    return userName, email, role, isVerified, nil
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
        (user_name, email, hash_password, profile_image_url, role, is_verified)
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
	var isVerified bool

	query := `
        SELECT user_id, user_name, hash_password, role, is_verified
        FROM trx_users 
        WHERE email = ?
        LIMIT 1
    `
	err := r.DB.QueryRow(query, email).Scan(&userID, &userName, &hashedPassword, &role, &isVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", "", "", false, fmt.Errorf("user not found")
		}
		return 0, "", "", "", false, fmt.Errorf("error retrieving user: %v", err)
	}

	return userID, userName, hashedPassword, role, isVerified, nil
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
