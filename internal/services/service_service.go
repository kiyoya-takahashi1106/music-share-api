package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"music-share-api/internal/repositories"
)

type SpotifyService interface {
	ConnectSpotify(userID int, code string) error
}

type spotifyService struct {
	repo repositories.ServiceRepository
}

func NewSpotifyService(repo repositories.ServiceRepository) SpotifyService {
	return &spotifyService{repo: repo}
}

// SpotifyTokenResponse はSpotifyのトークンレスポンスを表します。
type SpotifyTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// SpotifyUserProfile はSpotifyのユーザー情報レスポンスを表します。
type SpotifyUserProfile struct {
	ID string `json:"id"`
	// 必要に応じて追加フィールドを定義可能
}

// ConnectSpotify は、Spotifyの認証コード(code)を使用して
// 外部APIを呼び出し、アクセストークン、リフレッシュトークン、有効期限、SpotifyユーザーIDを取得しDBへ保存します。
func (s *spotifyService) ConnectSpotify(userID int, code string) error {
	if code == "" {
		return errors.New("code is empty")
	}

	// 必要な環境変数の取得
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	if clientID == "" || clientSecret == "" || redirectURI == "" {
		return errors.New("missing spotify credentials")
	}

	// Spotifyのトークンエンドポイント
	tokenURL := "https://accounts.spotify.com/api/token"
	formData := fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", code, redirectURI)
	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(formData))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Basic認証ヘッダーの設定
	authStr := fmt.Sprintf("%s:%s", clientID, clientSecret)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))
	req.Header.Set("Authorization", "Basic "+encodedAuth)

	// codeからアクセストークン、リフレッシュトークンを取得
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("token request failed, status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	userURL := "https://api.spotify.com/v1/me"
	reqUser, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create spotify user request: %w", err)
	}
	reqUser.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	// アクセストークンからspotifyユーザー情報を取得
	userResp, err := httpClient.Do(reqUser)
	if err != nil {
		return fmt.Errorf("failed to execute spotify user request: %w", err)
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(userResp.Body)
		return fmt.Errorf("spotify user request failed, status %d: %s", userResp.StatusCode, string(body))
	}

	var userProfile SpotifyUserProfile
	if err := json.NewDecoder(userResp.Body).Decode(&userProfile); err != nil {
		return fmt.Errorf("failed to decode spotify user response: %w", err)
	}

	// 有効期限の計算
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// 暗号化処理が必要な場合はここで実施
	encryptedAccessToken := tokenResp.AccessToken
	encryptedRefreshToken := tokenResp.RefreshToken

	// DBへ保存
	if err := s.repo.InsertUserService(userID, "spotify", userProfile.ID, encryptedAccessToken, encryptedRefreshToken, expiresAt); err != nil {
		return fmt.Errorf("failed to insert spotify service data: %w", err)
	}

	return nil
}
