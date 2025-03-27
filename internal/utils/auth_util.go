package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	return secret
}

var jwtSecret = []byte(getJWTSecret())

// 指定された http.ResponseWriter に JWT を含む httpOnly クッキーをセットします。
func SetAuthCookie(w http.ResponseWriter, userID int) error {
	claims := jwt.MapClaims{
		"iss":    "my-auth-server", // 発行者
		"userId": userID,
		"iat":    time.Now().Unix(), // 発行時間
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return fmt.Errorf("failed to sign JWT: %v", err)
	}

	cookie := &http.Cookie{
		Name:     "jwt_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true, // HTTPS環境の場合は true にしてください
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
	return nil
}


// cookieの有効期限の確認をし、userIDを返す。
func CheckAuthCookie(r *http.Request) (int, error) {
	// jwt_token クッキーを取得
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		log.Println("No jwt_token cookie found:", err)
		return 0, fmt.Errorf("no jwt_token cookie found")
	}
	log.Printf("jwt_token cookie: %s\n", cookie.Value)

	// JWT をパース
	tokenString := cookie.Value
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		log.Println("Failed to parse JWT:", err)
		return 0, fmt.Errorf("failed to parse JWT: %v", err)
	}

	// トークンが無効な場合
	if !token.Valid {
		log.Println("Invalid JWT token")
		return 0, fmt.Errorf("token is not valid")
	}

	// 有効期限 (exp) を確認
	expVal, ok := claims["exp"].(float64)
	if !ok {
		log.Println("JWT does not have an expiration claim")
		return 0, fmt.Errorf("token does not have an expiration claim")
	}
	if time.Unix(int64(expVal), 0).Before(time.Now()) {
		log.Println("JWT token has expired")
		return 0, fmt.Errorf("token has expired")
	}

	// userId を取得
	userIDFloat, ok := claims["userId"].(float64)
	if !ok {
		log.Println("Failed to get userId from JWT")
		return 0, fmt.Errorf("failed to get userId from JWT")
	}
	userID := int(userIDFloat)
	log.Printf("Authenticated user ID: %d\n", userID)

	log.Println("userId", userID)

	return userID, nil
}


func ClearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  time.Unix(0, 0), // 過去の日付を設定してクッキーを無効化
		HttpOnly: true,
		Secure:   true, // HTTPS環境の場合は true にしてください
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}