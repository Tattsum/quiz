// Package main provides the entry point for the quiz application server.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/handlers"
	"github.com/Tattsum/quiz/internal/middleware"
	"github.com/Tattsum/quiz/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// データベース接続初期化
	db, err := database.Initialize()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	// JWT サービス初期化
	jwtService := services.NewJWTService()

	// Ginルーターの設定
	router := gin.Default()

	// ミドルウェア設定
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimit())

	// API v1グループ
	v1 := router.Group("/api")

	// 認証エンドポイント（public）
	auth := v1.Group("/auth")
	{
		auth.POST("/login", handlers.AdminLogin)
		auth.POST("/refresh", handlers.RefreshToken)
	}

	// 管理者エンドポイント（認証が必要）
	admin := v1.Group("/admin")
	admin.Use(middleware.JWTAuth(jwtService))
	{
		// 認証関連
		admin.POST("/logout", handlers.AdminLogout)
		admin.GET("/verify", handlers.VerifyToken)
		
		// 問題管理
		admin.GET("/quizzes", handlers.GetQuizzes)
		admin.GET("/quizzes/:id", handlers.GetQuiz)
		admin.POST("/quizzes", handlers.CreateQuiz)
		admin.PUT("/quizzes/:id", handlers.UpdateQuiz)
		admin.DELETE("/quizzes/:id", handlers.DeleteQuiz)
		
		// セッション管理
		admin.POST("/session/start", handlers.StartSession)
		admin.POST("/session/next", handlers.NextQuestion)
		admin.POST("/session/toggle-answers", handlers.ToggleAnswers)
		admin.POST("/session/end", handlers.EndSession)
		
		// ファイルアップロード
		admin.POST("/upload/image", handlers.UploadImage)
		
		// 結果・ランキング（具体的なパスを先に定義）
		admin.GET("/results/current", handlers.GetCurrentResults)
		admin.GET("/ranking/overall", handlers.GetOverallRanking)
		admin.GET("/results/quiz/:id", handlers.GetQuizResults)
		admin.GET("/ranking/quiz/:id", handlers.GetQuizRanking)
		admin.GET("/ranking/participant/:id", handlers.GetParticipantRanking)
	}

	// セッション状態取得（公開）
	v1.GET("/session/status", handlers.GetSessionStatus)

	// 参加者関連エンドポイント
	participants := v1.Group("/participants")
	{
		participants.POST("/register", handlers.RegisterParticipant)
		participants.GET("/:id", handlers.GetParticipant)
		participants.GET("/:id/answers", handlers.GetParticipantAnswers)
	}

	// 回答関連エンドポイント
	answers := v1.Group("/answers")
	{
		answers.POST("", handlers.SubmitAnswer)
		answers.PUT("/:id", handlers.UpdateAnswer)
	}

	// 集計結果エンドポイント
	results := v1.Group("/results")
	{
		results.GET("/current", handlers.GetCurrentResults)
		results.GET("/quiz/:id", handlers.GetQuizResults)
	}

	// ランキングエンドポイント
	ranking := v1.Group("/ranking")
	{
		ranking.GET("/overall", handlers.GetOverallRanking)
		ranking.GET("/quiz/:id", handlers.GetQuizRanking)
		ranking.GET("/participant/:id", handlers.GetParticipantRanking)
	}

	// WebSocketエンドポイント
	v1.GET("/ws/results", handlers.WebSocketResults)

	// 静的ファイル配信 (アップロードされた画像など)
	router.Static("/uploads", "./uploads")

	// サーバー起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
