package middlewares

import (
	"net/http"

	"music-share-api/internal/utils"

	"github.com/gin-gonic/gin"
)


func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, err := utils.CheckAuthCookie(c.Request)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": err.Error()})
            c.Abort()
            return
        }

        // 取得した userID をコンテキストに保存して後続のハンドラーで利用可能にする
        c.Set("userID", userID)
        c.Next()
    }
}