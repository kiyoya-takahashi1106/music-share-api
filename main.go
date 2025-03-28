package main

import (
	"fmt"
	"log"
	"os"

	"music-share-api/internal/controllers"
	"music-share-api/internal/repositories"
	"music-share-api/internal/services"
	"music-share-api/internal/middlewares"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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

	// Redisクライアントを初期化
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redisサーバーのアドレス (環境に合わせて変更)
		Password: "",               // パスワード (必要であれば)
		DB:       0,                // デフォルトのDB
	})

	// リポジトリ、サービス、コントローラのセットアップ
	authRepository := repositories.NewAuthRepository(db.DB)
	authService := services.NewAuthService(authRepository)
	authController := controllers.NewAuthController(authService)

	roomsRepository := repositories.NewRoomsRepository(db.DB)
	roomsService := services.NewRoomsService(roomsRepository)
	roomsController := controllers.NewRoomsController(roomsService)

	// room作成用のセットアップ (Redisクライアントを追加)
	roomRepository := repositories.NewRoomRepository(db.DB, redisClient)
	roomService := services.NewRoomService(roomRepository)
	roomController := controllers.NewRoomController(roomService)

	// Ginルーター設定
	r := gin.Default()

	// CORS設定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		// AllowOrigins:  []string{"*"},
		AllowCredentials: true,
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Cookie", "Authorization", "Set-Cookie"},
		ExposeHeaders: []string{"Content-Length", "Set-Cookie"},
		MaxAge: 12 * time.Hour,
	}))

	// auth
	r.GET("/auth/user-info", middlewares.AuthMiddleware(), authController.GetUserInfo)	
	r.POST("/auth/sign-up", authController.SignUp)
	r.POST("/auth/sign-in", authController.SignIn)
	r.DELETE("/auth/sign-out", authController.SignOut)
	r.PUT("/auth/update-profile", authController.UpdateProfile)

	// spotify
	// r.GET("/spotify/auth", .SpotifyAuth)

	// rooms
	r.GET("/rooms/public", middlewares.AuthMiddleware(), roomsController.GetPublicRooms)

	// room
	r.POST("/room/create", roomController.CreateRoom)
	r.POST("/room/join", middlewares.AuthMiddleware(), roomController.JoinRoom)
	r.POST("/room/leave", roomController.LeaveRoom)
	r.DELETE("/room/delete/:roomId", roomController.DeleteRoom)
	r.GET("/room/:roomId", roomController.GetRoom)

	// サーバー起動
	r.Run(":8080")
}
