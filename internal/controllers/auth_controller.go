package controllers

import (
	"net/http"

	"music-share-api/internal/services"

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

// SignUp 新規ユーザー登録
func (ctrl *AuthController) SignUp(c *gin.Context) {
	var requestBody struct {
		UserName string `json:"user_name"`
		Email    string `json:"email"`
		Password string `json:"password"` // ハッシュ化前の平文パスワード
	}
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

	userID, userName, email, role, isVerified, err := ctrl.authService.LoginUser(
		requestBody.Email,
		requestBody.Password,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Invalid credentials"})
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
