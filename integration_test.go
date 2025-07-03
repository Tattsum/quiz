package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/handlers"
	"github.com/Tattsum/quiz/internal/middleware"
	"github.com/Tattsum/quiz/internal/models"
	"github.com/Tattsum/quiz/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	testDB     *sql.DB
	testRouter *gin.Engine
)

func TestMain(m *testing.M) {
	// テスト用の環境変数を設定
	os.Setenv("ENV", "test")
	os.Setenv("DATABASE_URL", "postgres://postgres:password@localhost:5432/quiz_test?sslmode=disable")
	os.Setenv("JWT_SECRET", "test_secret_key_for_testing_only")

	// .envファイルを読み込み（テスト環境では無視される場合がある）
	_ = godotenv.Load()

	// テスト用データベースに接続
	var err error
	testDB, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// データベース接続を確認
	if err = testDB.Ping(); err != nil {
		fmt.Printf("Failed to ping test database: %v\n", err)
		os.Exit(1)
	}

	// データベースを初期化
	setupTestDatabase()

	// テスト用ルーターを設定
	gin.SetMode(gin.TestMode)
	testRouter = setupTestRouter()

	// テストを実行
	code := m.Run()

	// クリーンアップ
	teardownTestDatabase()
	testDB.Close()

	os.Exit(code)
}

func setupTestDatabase() {
	// テーブルをクリア
	tables := []string{"answers", "quiz_sessions", "participants", "quizzes", "administrators"}
	for _, table := range tables {
		_, _ = testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}

	// テスト用管理者を作成
	_, err := testDB.Exec(`
		INSERT INTO administrators (username, password_hash, email, created_at, updated_at)
		VALUES ('testadmin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'test@example.com', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		fmt.Printf("Failed to create test admin: %v\n", err)
	}

	// テスト用クイズを作成
	_, err = testDB.Exec(`
		INSERT INTO quizzes (question_text, option_a, option_b, option_c, option_d, correct_answer, created_at, updated_at)
		VALUES 
		('What is 2+2?', '3', '4', '5', '6', 'B', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		('What is the capital of Japan?', 'Tokyo', 'Osaka', 'Kyoto', 'Nagoya', 'A', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		fmt.Printf("Failed to create test quizzes: %v\n", err)
	}

	// テスト用参加者を作成
	_, err = testDB.Exec(`
		INSERT INTO participants (nickname, created_at)
		VALUES 
		('TestUser1', CURRENT_TIMESTAMP),
		('TestUser2', CURRENT_TIMESTAMP)
	`)
	if err != nil {
		fmt.Printf("Failed to create test participants: %v\n", err)
	}
}

func teardownTestDatabase() {
	tables := []string{"answers", "quiz_sessions", "participants", "quizzes", "administrators"}
	for _, table := range tables {
		_, _ = testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}
}

func setupTestRouter() *gin.Engine {
	// データベースを設定
	database.SetTestDB(testDB)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// API routes
	api := r.Group("/api")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.AdminLogin)
			auth.POST("/refresh", handlers.RefreshToken)
		}

		// Admin routes (protected)
		admin := api.Group("/admin")
		jwtService := services.NewJWTService()
		admin.Use(middleware.JWTAuth(jwtService))
		{
			admin.POST("/logout", handlers.AdminLogout)
			admin.GET("/verify", handlers.VerifyToken)

			// Quiz management
			admin.GET("/quizzes", handlers.GetQuizzes)
			admin.POST("/quizzes", handlers.CreateQuiz)
			admin.GET("/quizzes/:id", handlers.GetQuiz)
			admin.PUT("/quizzes/:id", handlers.UpdateQuiz)
			admin.DELETE("/quizzes/:id", handlers.DeleteQuiz)

			// Session management
			admin.POST("/session/start", handlers.StartSession)
			admin.POST("/session/next", handlers.NextQuestion)
			admin.POST("/session/toggle-answers", handlers.ToggleAnswers)
			admin.GET("/session/status", handlers.GetSessionStatus)

			// Results and rankings
			admin.GET("/results/:quiz_id", handlers.GetQuizResults)
			admin.GET("/ranking/overall", handlers.GetOverallRanking)
			admin.GET("/ranking/quiz/:quiz_id", handlers.GetQuizRanking)
		}

		// Public routes
		public := api.Group("/public")
		{
			public.POST("/participants", handlers.RegisterParticipant)
			public.GET("/participants/:id", handlers.GetParticipant)
			public.GET("/participants/:id/answers", handlers.GetParticipantAnswers)
			public.POST("/answers", handlers.SubmitAnswer)
			public.PUT("/answers/:id", handlers.UpdateAnswer)
			public.GET("/quiz/:id", handlers.GetQuiz)
			public.GET("/session/status", handlers.GetSessionStatus)
		}
	}

	// WebSocket endpoint
	r.GET("/ws", handlers.WebSocketResults)

	return r
}

func TestIntegrationQuizFlow(t *testing.T) {
	// 1. 管理者ログイン
	loginReq := models.LoginRequest{
		Username: "testadmin",
		Password: "password",
	}
	loginBody, _ := json.Marshal(loginReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed: %d", w.Code)
	}

	var loginResp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &loginResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}

	loginData, ok := loginResp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Failed to parse login response data")
	}
	token, ok := loginData["access_token"].(string)
	if !ok {
		t.Fatal("Failed to parse access token")
	}

	// 2. クイズ作成
	quizReq := models.QuizRequest{
		QuestionText:  "Integration test question?",
		OptionA:       "Option A",
		OptionB:       "Option B",
		OptionC:       "Option C",
		OptionD:       "Option D",
		CorrectAnswer: "B",
	}
	quizBody, _ := json.Marshal(quizReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/admin/quizzes", bytes.NewBuffer(quizBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create quiz failed: %d", w.Code)
	}

	var quizResp models.APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &quizResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal quiz response: %v", err)
	}

	quizData, ok := quizResp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Failed to parse quiz response data")
	}
	quizIDFloat, ok := quizData["id"].(float64)
	if !ok {
		t.Fatal("Failed to parse quiz ID")
	}
	quizID := int64(quizIDFloat)

	// 3. セッション開始
	sessionReq := models.SessionStartRequest{
		QuizID: quizID,
	}
	sessionBody, _ := json.Marshal(sessionReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/admin/session/start", bytes.NewBuffer(sessionBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Start session failed: %d", w.Code)
	}

	// 4. 参加者登録
	participantReq := models.ParticipantRequest{
		Nickname: "IntegrationTestUser",
	}
	participantBody, _ := json.Marshal(participantReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/public/participants", bytes.NewBuffer(participantBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Register participant failed: %d", w.Code)
	}

	var participantResp models.APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &participantResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal participant response: %v", err)
	}

	participantData, ok := participantResp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Failed to parse participant response data")
	}
	participantIDFloat, ok := participantData["participant_id"].(float64)
	if !ok {
		t.Fatal("Failed to parse participant ID")
	}
	participantID := int64(participantIDFloat)

	// 5. 回答送信
	answerReq := models.AnswerRequest{
		ParticipantID:  participantID,
		QuizID:         quizID,
		SelectedOption: "B",
	}
	answerBody, _ := json.Marshal(answerReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/public/answers", bytes.NewBuffer(answerBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Submit answer failed: %d", w.Code)
	}

	// 6. 回答状況確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/admin/results/%d", quizID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get results failed: %d", w.Code)
	}

	var resultsResp models.APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &resultsResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal results response: %v", err)
	}

	resultsData := resultsResp.Data.(map[string]interface{})
	if resultsData["total_answers"].(float64) != 1 {
		t.Errorf("Expected 1 answer, got %v", resultsData["total_answers"])
	}

	if resultsData["correct_count"].(float64) != 1 {
		t.Errorf("Expected 1 correct answer, got %v", resultsData["correct_count"])
	}

	// 7. ランキング確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/admin/ranking/overall", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get ranking failed: %d", w.Code)
	}

	var rankingResp models.APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &rankingResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal ranking response: %v", err)
	}

	rankingData := rankingResp.Data.(map[string]interface{})
	if rankingData["total_participants"].(float64) < 1 {
		t.Errorf("Expected at least 1 participant in ranking, got %v", rankingData["total_participants"])
	}
}

func TestIntegrationSessionManagement(t *testing.T) {
	// 管理者ログイン
	loginReq := models.LoginRequest{
		Username: "testadmin",
		Password: "password",
	}
	loginBody, _ := json.Marshal(loginReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	var loginResp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	loginData := loginResp.Data.(map[string]interface{})
	token := loginData["access_token"].(string)

	// セッション状況確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/admin/session/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get session status failed: %d", w.Code)
	}

	// 回答受付切り替え
	toggleReq := models.ToggleAnswersRequest{
		IsAcceptingAnswers: false,
	}
	toggleBody, _ := json.Marshal(toggleReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/admin/session/toggle-answers", bytes.NewBuffer(toggleBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Toggle answers failed: %d", w.Code)
	}

	// 回答受付状況再確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/admin/session/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get session status after toggle failed: %d", w.Code)
	}

	var statusResp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &statusResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal status response: %v", err)
	}

	statusData := statusResp.Data.(map[string]interface{})
	if statusData["is_accepting_answers"].(bool) != false {
		t.Error("Expected answers to be disabled, but they are still enabled")
	}
}

func TestIntegrationParticipantFlow(t *testing.T) {
	// 参加者登録
	participantReq := models.ParticipantRequest{
		Nickname: "FlowTestUser",
	}
	participantBody, _ := json.Marshal(participantReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/public/participants", bytes.NewBuffer(participantBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Register participant failed: %d", w.Code)
	}

	var participantResp models.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &participantResp) // テスト用なのでエラーハンドリング不要
	participantData := participantResp.Data.(map[string]interface{})
	participantID := int64(participantData["participant_id"].(float64))

	// 参加者情報取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/public/participants/%d", participantID), nil)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get participant failed: %d", w.Code)
	}

	// 参加者回答履歴取得
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/public/participants/%d/answers", participantID), nil)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get participant answers failed: %d", w.Code)
	}

	var answersResp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &answersResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal answers response: %v", err)
	}

	// 初期状態では回答履歴は空
	answersData := answersResp.Data.(map[string]interface{})
	if answersData["total_answers"].(float64) != 0 {
		t.Errorf("Expected 0 answers for new participant, got %v", answersData["total_answers"])
	}
}

func TestIntegrationConcurrentAnswers(t *testing.T) {
	// 管理者でセッション開始
	loginReq := models.LoginRequest{
		Username: "testadmin",
		Password: "password",
	}
	loginBody, _ := json.Marshal(loginReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	var loginResp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	loginData := loginResp.Data.(map[string]interface{})
	token := loginData["access_token"].(string)

	// 最初のクイズでセッション開始
	sessionReq := models.SessionStartRequest{
		QuizID: 1,
	}
	sessionBody, _ := json.Marshal(sessionReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/admin/session/start", bytes.NewBuffer(sessionBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)

	// 複数参加者の同時回答をシミュレート
	numParticipants := 10
	done := make(chan bool, numParticipants)

	for i := 0; i < numParticipants; i++ {
		go func(userNum int) {
			// 参加者登録
			participantReq := models.ParticipantRequest{
				Nickname: fmt.Sprintf("ConcurrentUser%d", userNum),
			}
			participantBody, _ := json.Marshal(participantReq)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/public/participants", bytes.NewBuffer(participantBody))
			req.Header.Set("Content-Type", "application/json")
			testRouter.ServeHTTP(w, req)

			if w.Code == http.StatusCreated {
				var participantResp models.APIResponse
				_ = json.Unmarshal(w.Body.Bytes(), &participantResp) // テスト用なのでエラーハンドリング不要
				participantData := participantResp.Data.(map[string]interface{})
				participantID := int64(participantData["participant_id"].(float64))

				// 回答送信
				answerReq := models.AnswerRequest{
					ParticipantID:  participantID,
					QuizID:         1,
					SelectedOption: []string{"A", "B", "C", "D"}[userNum%4],
				}
				answerBody, _ := json.Marshal(answerReq)
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("POST", "/api/public/answers", bytes.NewBuffer(answerBody))
				req.Header.Set("Content-Type", "application/json")
				testRouter.ServeHTTP(w, req)
			}

			done <- true
		}(i)
	}

	// すべてのgoroutineの完了を待つ
	timeout := time.After(10 * time.Second)
	completed := 0

	for completed < numParticipants {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatalf("Timeout waiting for concurrent operations to complete")
		}
	}

	// 結果を確認
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/admin/results/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get results after concurrent answers failed: %d", w.Code)
	}

	var resultsResp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resultsResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal results response: %v", err)
	}

	resultsData := resultsResp.Data.(map[string]interface{})
	totalAnswers := resultsData["total_answers"].(float64)

	if totalAnswers < float64(numParticipants) {
		t.Errorf("Expected at least %d answers, got %v", numParticipants, totalAnswers)
	}
}
