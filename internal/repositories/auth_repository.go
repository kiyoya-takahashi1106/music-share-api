package repositories

import (
    "database/sql"
    "fmt"
    "log"
)

type AuthRepository interface {
    CreateUser(userName, email, hashedPassword string) (int, error)
    GetUserByEmail(email string) (int, string, string, string, bool, error)
}

type authRepository struct {
    DB *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
    return &authRepository{DB: db}
}

func (r *authRepository) CreateUser(userName, email, hashedPassword string) (int, error) {
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