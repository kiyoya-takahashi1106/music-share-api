package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"music-share-api/internal/repositories"
)

type SpotifyService interface {
	ConnectSpotify(userID int, code string) error
	DeleteSpotify(userID int) error
	RefreshSpotifyToken(userID int) (string, time.Time, error)
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
// ここで display_name を追加して、アカウント名を取得します。
type SpotifyUserProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	// 必要に応じて追加フィールドを定義可能
}

// encodeToken は、トークンのエンコード（例: Base64エンコード）を実施するヘルパー関数です。
func encodeToken(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

// ConnectSpotify は、Spotifyの認証コード(code)を使用して
// 外部APIを呼び出し、アクセストークン、リフレッシュトークン、有効期限、
// SpotifyユーザーIDおよびアカウント名を取得しDBへ保存します。
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

	// Spotifyユーザー情報取得 (/v1/me)
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

	// JSTのタイムゾーンを取得
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return fmt.Errorf("failed to load JST location: %w", err)
	}
	// 有効期限を日本時刻で計算
	currentTime := time.Now().In(loc)
	expiresAt := currentTime.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// 暗号化処理が必要な場合はここで実施（今回はそのまま）
	encryptedAccessToken := encodeToken(tokenResp.AccessToken)
	encryptedRefreshToken := encodeToken(tokenResp.RefreshToken)

	// DBへ保存
	// InsertUserService の第4引数として Spotify のアカウント名 (DisplayName) を渡す
	if err := s.repo.InsertUserService(userID, "spotify", userProfile.ID, userProfile.DisplayName, encryptedAccessToken, encryptedRefreshToken, expiresAt); err != nil {
		return fmt.Errorf("failed to insert spotify service data: %w", err)
	}

	return nil
}


// DeleteSpotify は、Spotifyのサービスデータを削除します。
func (s *spotifyService) DeleteSpotify(userID int) error {
	// "spotify" を指定してレコードを削除
	if err := s.repo.DeleteUserService(userID, "spotify"); err != nil {
		return fmt.Errorf("failed to delete spotify service data: %w", err)
	}
	return nil
}


// RefreshSpotifyToken は、DBからエンコード済みの refresh token を取得し、デコード後にSpotify API を呼び出して
// 新しいアクセストークンを取得、エンコードしてDBを更新します。
func (s *spotifyService) RefreshSpotifyToken(userID int) (string, time.Time, error) {
	// DBからエンコード済みのリフレッシュトークンを取得
	encryptedStoredToken, err := s.repo.GetSpotifyRefreshToken(userID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to retrieve refresh token: %w", err)
	}

	// 取得したリフレッシュトークンをデコードする
	decodedBytes, err := base64.StdEncoding.DecodeString(encryptedStoredToken)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to decode stored refresh token: %w", err)
	}
	refreshToken := string(decodedBytes)

	// Spotifyのクライアント情報の取得
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", time.Time{}, errors.New("missing spotify credentials")
	}

	// トークンリフレッシュ用エンドポイント
	tokenURL := "https://accounts.spotify.com/api/token"
	// POSTパラメータ
	formData := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", refreshToken)
	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(formData))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create token refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	authStr := fmt.Sprintf("%s:%s", clientID, clientSecret)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))
	req.Header.Set("Authorization", "Basic "+encodedAuth)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to execute token refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", time.Time{}, fmt.Errorf("token refresh request failed, status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to decode refresh token response: %w", err)
	}

	// DB更新前の新しい有効期限を日本時刻で計算
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to load JST location: %w", err)
	}
	newExpiresAt := time.Now().In(loc).Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	log.Printf("Current JST time: %s", time.Now().In(loc).Format(time.RFC3339))
	log.Printf("expiresAt: %s", newExpiresAt.Format(time.RFC3339))

	// 通常はリフレッシュトークンは再発行されないが、新しい値が返ってくる場合もある
	newRefreshToken := refreshToken
	if tokenResp.RefreshToken != "" {
		newRefreshToken = tokenResp.RefreshToken
	}

	// ConnectSpotify と同様にエンコードしてから repository に入れる
	encryptedAccessToken := encodeToken(tokenResp.AccessToken)
	encryptedRefreshToken := encodeToken(newRefreshToken)

	// DBのトークン情報を更新する
	if err := s.repo.UpdateSpotifyToken(userID, encryptedAccessToken, encryptedRefreshToken, newExpiresAt); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to update spotify token in db: %w", err)
	}

	return encryptedAccessToken, newExpiresAt, nil
}
