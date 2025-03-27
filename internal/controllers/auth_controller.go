package controllers

import (
	"fmt"
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

// cookieからユーザー情報を取得
func (ctrl *AuthController) GetUserInfo(c *gin.Context) {
	// ミドルウェアでセットされた userID をコンテキストから取得
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
	userName, email, role, isVerified, err := ctrl.authService.GetUserInfo(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid credentials"})
		return
	}

	// JSON情報を返す
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "User information retrieved successfully",
		"user_id":     userID,
		"user_name":   userName,
		"email":       email,
		"role":        role,
		"is_verified": isVerified,
	})
}


// SignUp 新規ユーザー登録
func (ctrl *AuthController) SignUp(c *gin.Context) {
	var requestBody struct {
		UserName string `json:"user_name"`
		Email    string `json:"email"`
		Password string `json:"password"` // ハッシュ化前の平文パスワード
	}
	fmt.Printf("%+v\n", requestBody)

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid input"})
		return
	}

	userID, userName, email, err := ctrl.authService.RegisterUser(
		requestBody.UserName,
		requestBody.Email,
		requestBody.Password,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// HTTPレスポンスにクッキーをセット
	if err := utils.SetAuthCookie(c.Writer, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to set auth cookie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "User registered successfully",
		"user_id":   userID,
		"user_name": userName,
		"email":     email,
	})
}


// SignIn ログイン処理
func (ctrl *AuthController) SignIn(c *gin.Context) {
	var requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"` // ハッシュ化前の平文パスワード
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid input"})
		return
	}

	userID, userName, email, role, isVerified, err := ctrl.authService.LoginUser(requestBody.Email, requestBody.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid credentials"})
		return
	}

	// HTTPレスポンスにクッキーをセット
	if err := utils.SetAuthCookie(c.Writer, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to set auth cookie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "Login successful",
		"user_id":     userID,
		"user_name":   userName,
		"email":       email,
		"role":        role,
		"is_verified": isVerified,
	})
}


// サインアウト
func (ctrl *AuthController) SignOut(c *gin.Context) {
	// クッキーを削除
	utils.ClearAuthCookie(c.Writer)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Sign out successful",
	})
}


// profile更新
func (ctrl *AuthController) UpdateProfile(c *gin.Context) {
	// リクエストボディのデータをバインドする構造体
	var req struct {
		UserID   int    `json:"user_id" binding:"required"`
		UserName string `json:"user_name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid input",
		})
		return
	}

	// サービス経由でプロファイル更新処理
	if err := ctrl.authService.UpdateProfile(req.UserID, req.UserName, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// 指示された通り、更新完了時は以下のJSONを返す
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Account deleted successfully",
	})
}
