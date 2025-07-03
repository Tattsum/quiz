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

	// Ginルーターの設定
	router := gin.Default()

	// ミドルウェア設定
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimit())

	// API v1グループ
	v1 := router.Group("/api")

	// 管理者認証エンドポイント
	adminAuth := v1.Group("/admin")
	{
		adminAuth.POST("/login", handlers.AdminLogin)
		adminAuth.POST("/logout", middleware.JWTAuth(), handlers.AdminLogout)
		adminAuth.GET("/verify", middleware.JWTAuth(), handlers.VerifyToken)
	}

	// 管理者用問題管理エンドポイント
	adminQuiz := v1.Group("/admin/quizzes")
	adminQuiz.Use(middleware.JWTAuth())
	{
		adminQuiz.GET("", handlers.GetQuizzes)
		adminQuiz.GET("/:id", handlers.GetQuiz)
		adminQuiz.POST("", handlers.CreateQuiz)
		adminQuiz.PUT("/:id", handlers.UpdateQuiz)
		adminQuiz.DELETE("/:id", handlers.DeleteQuiz)
	}

	// 管理者用セッション管理エンドポイント
	adminSession := v1.Group("/admin/session")
	adminSession.Use(middleware.JWTAuth())
	{
		adminSession.POST("/start", handlers.StartSession)
		adminSession.POST("/next", handlers.NextQuestion)
		adminSession.POST("/toggle-answers", handlers.ToggleAnswers)
		adminSession.POST("/end", handlers.EndSession)
	}

	// 管理者用ファイルアップロード
	adminUpload := v1.Group("/admin/upload")
	adminUpload.Use(middleware.JWTAuth())
	{
		adminUpload.POST("/image", handlers.UploadImage)
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
