package services

import (
	"errors"
	"music-share-api/internal/repositories"
)

type AuthService interface {
	// 新規：ユーザー基本情報と連携サービス情報を返す
	GetUserInfo(userID int) (string, string, string, bool, map[string]repositories.UserServiceData, error)
	RegisterUser(userName, email, hashPassword string) (int, string, string, error)
	LoginUser(email, hashPassword string) (int, string, string, string, bool, error)
	UpdateProfile(userID int, userName, email string) error
}

type authService struct {
	repo repositories.AuthRepository
}

func NewAuthService(r repositories.AuthRepository) AuthService {
	return &authService{repo: r}
}

func (s *authService) GetUserInfo(userID int) (string, string, string, bool, map[string]repositories.UserServiceData, error) {
	return s.repo.GetUserInfo(userID)
}

// RegisterUser：パスワードは既にハッシュ化されている前提
func (s *authService) RegisterUser(userName, email, hashPassword string) (int, string, string, error) {
	id, err := s.repo.CreateUser(userName, email, hashPassword)
	if err != nil {
		return 0, "", "", err
	}
	return id, userName, email, nil
}

// LoginUser：受け取ったhashPasswordとDBのものを直接比較
func (s *authService) LoginUser(email, hashPassword string) (int, string, string, string, bool, error) {
	id, name, storedHash, role, isSpotify, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return 0, "", "", "", false, err
	}
	if storedHash != hashPassword {
		return 0, "", "", "", false, errors.New("invalid credentials")
	}
	return id, name, email, role, isSpotify, nil
}

func (s *authService) UpdateProfile(userID int, userName, email string) error {
	return s.repo.UpdateUserProfile(userID, userName, email)
}
