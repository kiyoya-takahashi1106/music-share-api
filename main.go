package main

import (
	"fmt"
	"log"
	"os"

	"music-share-api/internal/controllers"
	"music-share-api/internal/repositories"
	"music-share-api/internal/services"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルから環境変数を読み込む
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// 環境変数からDB接続情報を取得
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)

	// データベース接続設定
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// リポジトリ、サービス、コントローラのセットアップ
	authRepository := repositories.NewAuthRepository(db.DB)
	authService := services.NewAuthService(authRepository)
	authController := controllers.NewAuthController(authService)

	roomsRepository := repositories.NewRoomsRepository(db.DB)
	roomsService := services.NewRoomsService(roomsRepository)
	roomsController := controllers.NewRoomsController(roomsService)

	// ルーム作成用のセットアップ
	roomRepository := repositories.NewRoomRepository(db.DB)
	roomService := services.NewRoomService(roomRepository)
	roomController := controllers.NewRoomController(roomService)

	// Ginルーター設定
	r := gin.Default()

	// CORSミドルウェアの追加
	r.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"http://localhost:3000"},
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders: []string{"Content-Length"},
		// AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))

	// ルーティング
	// auth
	r.POST("/auth/sign-up", authController.SignUp)
	r.POST("/auth/sign-in", authController.SignIn)

	// rooms
	r.GET("/rooms/public", roomsController.GetPublicRooms)

	// room作成
	r.POST("/room/create", roomController.CreateRoom)

	// サーバー起動
	r.Run(":8080")
}
