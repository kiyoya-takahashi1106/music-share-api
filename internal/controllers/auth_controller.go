package controllers

import (
    "net/http"

    "music-share-api/internal/services"

    "github.com/gin-gonic/gin"
)

// ServiceInfo は連携サービスの情報を表します
type ServiceInfo struct {
    ServiceName          string      `json:"service_name"`
    ServiceUserID        string      `json:"service_user_id"`
    EncryptedAccessToken string      `json:"encrypted_access_token"`
    ExpiresAt            interface{} `json:"expires_at"`
}

type AuthController struct {
    authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
    return &AuthController{
        authService: authService,
    }
}


// SignUp 新規ユーザー登録
func (ctrl *AuthController) SignUp(c *gin.Context) {
    var requestBody struct {
        UserName     string `json:"user_name"`
        Email        string `json:"email"`
        HashPassword string `json:"hash_password"`
    }
    if err := c.BindJSON(&requestBody); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid input"})
        return
    }

    userID, userName, email, err := ctrl.authService.RegisterUser(
        requestBody.UserName,
        requestBody.Email,
        requestBody.HashPassword,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
        return
    }

    // 仮のサービス情報（実際の実装では、ユーザー連携サービス情報を取得する）
    servicesInfo := []ServiceInfo{
        {
            ServiceName:          "spotify",
            ServiceUserID:        "sample_service_user_id",
            EncryptedAccessToken: "sample_encrypted_access_token",
            ExpiresAt:            nil, // 実際の有効期限日時を設定
        },
    }

    c.JSON(http.StatusOK, gin.H{
        "status":      "success",
        "message":     "User registered successfully",
        "user_id":     userID,
        "user_name":   userName,
        "email":       email,
        "role":        "user",
        "is_verified": "false",
        "services":    servicesInfo,
    })
}


// SignIn ログイン処理
func (ctrl *AuthController) SignIn(c *gin.Context) {
    var requestBody struct {
        Email        string `json:"email"`
        HashPassword string `json:"hash_password"`
    }
    if err := c.BindJSON(&requestBody); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid input"})
        return
    }

    userID, userName, email, role, isVerified, err := ctrl.authService.LoginUser(
        requestBody.Email,
        requestBody.HashPassword,
    )
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid credentials"})
        return
    }

    // 仮のサービス情報（実際の実装では、ユーザー連携サービス情報を取得する）
    servicesInfo := []ServiceInfo{
        {
            ServiceName:          "spotify",
            ServiceUserID:        "sample_service_user_id",
            EncryptedAccessToken: "sample_encrypted_access_token",
            ExpiresAt:            nil, // 実際の有効期限日時を設定
        },
    }

    c.JSON(http.StatusOK, gin.H{
        "status":      "success",
        "message":     "Login successful",
        "user_id":     userID,
        "user_name":   userName,
        "email":       email,
        "role":        role,
        "is_verified": isVerified,
        "services":    servicesInfo,
    })
}
