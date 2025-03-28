package controllers

import (
	"log"
	"net/http"

	"music-share-api/internal/services"
	"music-share-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// GET /auth/user-info
// フロントからはCookieでユーザーIDを取得し、DBの情報と固定のservicesオブジェクトを返す
func (ctrl *AuthController) GetUserInfo(c *gin.Context) {
	// ミドルウェアでセットされたユーザーIDを取得
	authUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}
	userID, ok := authUserID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to parse user ID",
		})
		return
	}

	// サービスからユーザー情報を取得
	userName, email, role, isSpotify, err := ctrl.authService.GetUserInfo(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid credentials"})
		return
	}

	// 要件に合わせた固定のservices情報を付与して返す
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "User registered successfully",
		"userId":    userID,
		"userName":  userName,
		"email":     email,
		"role":      role,
		"isSpotify": isSpotify,
		"services": gin.H{
			"spotify": gin.H{
				"serviceUserId":        "fgthrytjyhrgafegthryj",
				"encryptedAccessToken": "dwfagehjtythgrefe",
				"expiresAt":            false,
			},
		},
	})
}


// POST /auth/sign-up
// 新規ユーザー登録（フロントからは userName, email, hashPassword を受ける）
func (ctrl *AuthController) SignUp(c *gin.Context) {
	var requestBody struct {
		UserName     string `json:"userName"`
		Email        string `json:"email"`
		HashPassword string `json:"hashPassword"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid input"})
		return
	}

	// 登録処理（ここでは既にハッシュ化されたパスワードを使用）
	userID, userName, email, err := ctrl.authService.RegisterUser(
		requestBody.UserName,
		requestBody.Email,
		requestBody.HashPassword,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// クッキーのセット
	if err := utils.SetAuthCookie(c.Writer, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to set auth cookie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "User registered successfully",
		"userId":    userID,
		"userName":  userName,
		"email":     email,
		"role":      "user",
		"isSpotify": false,
		"services": gin.H{
			"spotify": gin.H{
				"serviceUserId":        "fgthrytjyhrgafegthryj",
				"encryptedAccessToken": "dwfagehjtythgrefe",
				"expiresAt":            false,
			},
		},
	})
}


// POST /auth/sign-in
// ログイン処理（フロントからは email と hashPassword を受ける）
func (ctrl *AuthController) SignIn(c *gin.Context) {
	var requestBody struct {
		Email        string `json:"email"`
		HashPassword string `json:"hashPassword"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid input"})
		return
	}

	userID, userName, email, role, isSpotify, err := ctrl.authService.LoginUser(requestBody.Email, requestBody.HashPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid credentials"})
		return
	}

	// クッキーのセット
	if err := utils.SetAuthCookie(c.Writer, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to set auth cookie"})
		return
	}

	log.Println("User ID:", userID)
	log.Println("User Name:", userName)
	log.Println("Email:", email)
	log.Println("Role:", role)
	log.Println("Is Spotify:", isSpotify)

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Login successful",
		"userId":    userID,
		"userName":  userName,
		"email":     email,
		"role":      role,
		"isSpotify": isSpotify,
		"services": gin.H{
			"spotify": gin.H{
				"serviceUserId":        "fgthrytjyhrgafegthryj",
				"encryptedAccessToken": "dwfagehjtythgrefe",
				"expiresAt":            false,
			},
		},
	})
}


// PUT /auth/update-profile
// プロファイル更新（※要件に合わせ、返すJSONは固定値とする）
func (ctrl *AuthController) UpdateProfile(c *gin.Context) {
	var req struct {
		UserId   int    `json:"userId" binding:"required"`
		UserName string `json:"userName" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input",
		})
		return
	}

	if err := ctrl.authService.UpdateProfile(req.UserId, req.UserName, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// 要件に合わせた固定のレスポンス
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Account deleted successfully",
	})
}


// POST /auth/sign-out
// ログアウト処理
func (ctrl *AuthController) SignOut(c *gin.Context) {
	// クッキーを削除
	utils.ClearAuthCookie(c.Writer)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logout successful",
	})
}
