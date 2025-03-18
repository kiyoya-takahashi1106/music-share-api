package services

import (
	"errors"
	"music-share-api/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	// RegisterUser 新規ユーザー登録
	RegisterUser(userName, email, hashedPassword string) (int, string, string, error)
	// LoginUser ログイン処理
	LoginUser(email, password string) (int, string, string, string, bool, error)
}

type authService struct {
	authRepo repositories.AuthRepository
}

func NewAuthService(authRepo repositories.AuthRepository) AuthService {
	return &authService{
		authRepo: authRepo,
	}
}

// RegisterUser 新規ユーザー登録
func (s *authService) RegisterUser(userName, email, hashedPassword string) (int, string, string, error) {
	userID, err := s.authRepo.CreateUser(userName, email, hashedPassword)
	if err != nil {
		return 0, "", "", err
	}
	return userID, userName, email, nil
}


// LoginUser ログイン処理
func (s *authService) LoginUser(email, password string) (int, string, string, string, bool, error) {
	userID, userName, hashedPassword, role, isVerified, err := s.authRepo.GetUserByEmail(email)
	if err != nil {
		return 0, "", "", "", false, err
	}

	// パスワード検証
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return 0, "", "", "", false, errors.New("invalid credentials")
	}

	return userID, userName, email, role, isVerified, nil
}
