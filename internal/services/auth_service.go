package services

import (
	"errors"
	"fmt"
	"music-share-api/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	GetUserInfo(userID int) (string, string, string, bool, error)
	RegisterUser(userName, email, password string) (int, string, string, error)
	LoginUser(email, password string) (int, string, string, string, bool, error)
	UpdateProfile(userID int, userName, email string) error
}

type authService struct {
	repo repositories.AuthRepository
}

func NewAuthService(r repositories.AuthRepository) AuthService {
	return &authService{repo: r}
}


func (s *authService) GetUserInfo(userID int) (string, string, string, bool, error) {
    userName, email, role, isVerified, err := s.repo.GetUserInfo(userID)
    if err != nil {
        return "", "", "", false, err
    }
    return userName, email, role, isVerified, nil
}


// RegisterUser 新規ユーザー登録
func (s *authService) RegisterUser(userName, email, password string) (int, string, string, error) {
	// パスワードをハッシュ化
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", "", err
	}
	hashed := string(hashedBytes)

	id, err := s.repo.CreateUser(userName, email, hashed)
	if err != nil {
		return 0, "", "", err
	}
	return id, userName, email, nil
}

// LoginUser ログイン処理
func (s *authService) LoginUser(email, password string) (int, string, string, string, bool, error) {
	id, name, hashed, role, verified, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return 0, "", "", "", false, err
	}

	// 受け取った平文のパスワードと、DBのハッシュ値を比較します
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return 0, "", "", "", false, errors.New("invalid credentials")
	}

	return id, name, email, role, verified, nil
}

func (s *authService) UpdateProfile(userID int, userName, email string) error {
	if err := s.repo.UpdateUserProfile(userID, userName, email); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}
	return nil
}
