package repositories

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/go-sql-driver/mysql"
)

type AuthRepository interface {
    CreateUser(userName, email, hashedPassword, profileImageURL, role string) (int, error)
    GetUserByEmail(email string) (int, string, string, string, bool, error)
}

type authRepository struct {
    DB *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
    return &authRepository{
        DB: db,
    }
}

// CreateUser 新規ユーザー作成
func (r *authRepository) CreateUser(userName, email, hashedPassword) (int, error) {
    query := `
				INSERT INTO trx_users 
				(user_name, email, hash_password, profile_image_url, role, is_verified)
              	VALUES (?, ?, ?, ?, ?, ?)
			 `

    result, err := r.DB.Exec(query, userName, email, hashedPassword)
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


// GetUserByEmail メールでユーザーを取得
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
