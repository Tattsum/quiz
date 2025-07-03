// Package main provides performance tests for the quiz application.
//
// パフォーマンステスト実行前の要件:
// 1. サーバーがポート8080で起動していること
// 2. PostgreSQLデータベースが起動していること
// 3. テスト用環境変数が設定されていること
//
// 実行方法:
//
//	RUN_PERFORMANCE_TESTS=true go test -v -run TestConcurrent -timeout 15m ./performance_test.go
//
// テスト内容:
// - 70人同時参加者登録テスト
// - 70人同時WebSocket接続テスト
// - 70人同時回答送信テスト
// - システム全体負荷テスト
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Tattsum/quiz/internal/models"
	"github.com/gorilla/websocket"
)

const (
	// パフォーマンステスト用の設定
	MaxConcurrentUsers = 70
	runPerfTestsEnv    = "true"
	BaseURL            = "http://localhost:8080"
	WebSocketURL       = "ws://localhost:8080/ws"
	TestDuration       = 30 * time.Second
	// タイムアウト設定（GitHub Actions環境を考慮）
	HTTPTimeout      = 15 * time.Second // GitHub Actions環境向けに延長
	WebSocketTimeout = 15 * time.Second // GitHub Actions環境向けに延長
)

// パフォーマンステストの結果を記録する構造体
type PerformanceResult struct {
	TotalRequests  int
	SuccessfulReqs int
	FailedRequests int
	AverageLatency time.Duration
	MaxLatency     time.Duration
	MinLatency     time.Duration
	RequestsPerSec float64
	ErrorRate      float64
}

// 単一リクエストの結果
type RequestResult struct {
	Success   bool
	Latency   time.Duration
	Error     error
	Timestamp time.Time
}

// HTTPクライアントにタイムアウトを設定するヘルパー関数
func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: HTTPTimeout,
	}
}

// WebSocketダイアラーにタイムアウトを設定するヘルパー関数
func createWebSocketDialer() *websocket.Dialer {
	return &websocket.Dialer{
		HandshakeTimeout: WebSocketTimeout,
	}
}

// GitHub Actions環境での設定を取得するヘルパー関数
func getMaxConcurrentUsers() int {
	// GitHub Actions環境では少し控えめに設定
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return 50 // GitHub Actions環境では70から50に削減
	}
	return MaxConcurrentUsers
}

func getTestDuration() time.Duration {
	// GitHub Actions環境では短縮
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return 20 * time.Second // 30秒から20秒に短縮
	}
	return TestDuration
}

// テスト前のセットアップを行う
func setupPerformanceTest(t *testing.T) {
	// 1. サーバーが起動していることを確認
	t.Helper()
	t.Log("サーバーのヘルスチェックを実行中...")
	client := createHTTPClient()
	resp, err := client.Get(BaseURL + "/api/session/status")
	if err != nil || resp == nil {
		t.Fatalf("サーバーが起動していません。以下を確認してください:\n"+
			"1. サーバーが %s で起動していること\n"+
			"2. データベースが起動していること\n"+
			"エラー: %v", BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("サーバーのヘルスチェックに失敗しました: HTTP status %d", resp.StatusCode)
	}

	// 2. WebSocketエンドポイントの確認
	t.Log("WebSocketエンドポイントの確認中...")
	dialer := createWebSocketDialer()
	wsConn, _, err := dialer.Dial(WebSocketURL, nil)
	if err != nil {
		t.Fatalf("WebSocket接続に失敗しました: %v", err)
	}
	wsConn.Close()

	// 3. データベース接続の間接的確認（参加者登録テスト）
	t.Log("データベース接続の確認中...")
	testParticipant := models.ParticipantRequest{
		Nickname: "HealthCheckUser",
	}
	jsonData, _ := json.Marshal(testParticipant)
	testResp, err := client.Post(
		BaseURL+"/api/participants/register",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil || testResp == nil {
		t.Fatalf("データベース接続確認に失敗しました: %v", err)
	}
	defer testResp.Body.Close()

	if testResp.StatusCode != http.StatusCreated {
		t.Fatalf("テスト用参加者登録に失敗しました: HTTP status %d", testResp.StatusCode)
	}

	// 4. 管理者ログインと認証トークン取得
	t.Log("管理者認証中...")
	loginReq := models.LoginRequest{
		Username: "admin",
		Password: "password",
	}
	loginData, _ := json.Marshal(loginReq)
	loginResp, err := client.Post(
		BaseURL+"/api/auth/login",
		"application/json",
		bytes.NewBuffer(loginData),
	)
	if err != nil || loginResp == nil {
		t.Fatalf("管理者ログインに失敗しました: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("管理者認証に失敗しました: HTTP status %d", loginResp.StatusCode)
	}

	var loginResult map[string]interface{}
	if err := json.NewDecoder(loginResp.Body).Decode(&loginResult); err != nil {
		t.Fatalf("ログインレスポンスの解析に失敗しました: %v", err)
	}

	loginResultData, ok := loginResult["data"].(map[string]interface{})
	if !ok {
		t.Fatal("ログインレスポンスのデータ形式が不正です")
	}
	token, ok := loginResultData["access_token"].(string)
	if !ok {
		t.Fatal("アクセストークンの取得に失敗しました")
	}

	// 5. テスト用クイズID 2の存在確認と作成
	t.Log("テスト用クイズの準備中...")
	req, _ := http.NewRequest("GET", BaseURL+"/api/admin/quizzes/2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	quizResp, err := client.Do(req)
	
	if err != nil || quizResp.StatusCode != http.StatusOK {
		// クイズID 2が存在しない場合は作成
		t.Log("テスト用クイズを作成中...")
		quizReq := models.QuizRequest{
			QuestionText:  "パフォーマンステスト用問題",
			OptionA:       "選択肢A",
			OptionB:       "選択肢B", 
			OptionC:       "選択肢C",
			OptionD:       "選択肢D",
			CorrectAnswer: "A",
		}
		quizData, _ := json.Marshal(quizReq)
		req, _ := http.NewRequest("POST", BaseURL+"/api/admin/quizzes", bytes.NewBuffer(quizData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		
		createResp, err := client.Do(req)
		if err != nil || createResp == nil || createResp.StatusCode != http.StatusCreated {
			t.Fatalf("テスト用クイズの作成に失敗しました: %v", err)
		}
		createResp.Body.Close()
	} else {
		quizResp.Body.Close()
	}

	// 6. セッション開始
	t.Log("パフォーマンステスト用セッションを開始中...")
	sessionReq := models.SessionStartRequest{
		QuizID: 2,
	}
	sessionData, _ := json.Marshal(sessionReq)
	req, _ = http.NewRequest("POST", BaseURL+"/api/admin/session/start", bytes.NewBuffer(sessionData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	sessionResp, err := client.Do(req)
	if err != nil || sessionResp == nil {
		t.Fatalf("セッション開始に失敗しました: %v", err)
	}
	defer sessionResp.Body.Close()

	if sessionResp.StatusCode != http.StatusOK {
		t.Fatalf("セッション開始に失敗しました: HTTP status %d", sessionResp.StatusCode)
	}

	t.Log("✅ パフォーマンステスト環境のセットアップが完了しました")
	t.Logf("  - API Base URL: %s", BaseURL)
	t.Logf("  - WebSocket URL: %s", WebSocketURL)
	t.Logf("  - 最大同時ユーザー数: %d", getMaxConcurrentUsers())
	t.Logf("  - テスト継続時間: %v", getTestDuration())
	t.Logf("  - テスト用クイズID: 2")
	t.Logf("  - セッション状態: アクティブ（回答受付中）")
}

// テスト後のクリーンアップを行う
func cleanupPerformanceTest(t *testing.T) {
	t.Helper()
	t.Log("パフォーマンステストのクリーンアップを実行中...")

	// WebSocket接続の最終確認（サーバーが正常に動作していることを確認）
	dialer := createWebSocketDialer()
	wsConn, _, err := dialer.Dial(WebSocketURL, nil)
	if err != nil {
		t.Logf("警告: クリーンアップ時のWebSocket接続に失敗: %v", err)
	} else {
		wsConn.Close()
		t.Log("  - WebSocket接続が正常に動作しています")
	}

	// サーバーの最終ヘルスチェック
	client := createHTTPClient()
	resp, err := client.Get(BaseURL + "/api/session/status")
	if err != nil {
		t.Logf("警告: クリーンアップ時のサーバーヘルスチェックに失敗: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Log("  - サーバーが正常に動作しています")
		} else {
			t.Logf("警告: サーバーのレスポンスが異常です: status %d", resp.StatusCode)
		}
	}

	// パフォーマンステストでは、次回実行のために基本データは残す
	// テスト用参加者データの大量削除は行わない（パフォーマンスに影響するため）

	t.Log("✅ パフォーマンステストのクリーンアップが完了しました")
}

func TestConcurrentParticipantRegistration(t *testing.T) {
	if testing.Short() && os.Getenv("RUN_PERFORMANCE_TESTS") != runPerfTestsEnv {
		t.Skip("Skipping performance test in short mode")
	}

	setupPerformanceTest(t)
	defer cleanupPerformanceTest(t)

	numParticipants := getMaxConcurrentUsers()
	results := make(chan RequestResult, numParticipants)
	var wg sync.WaitGroup

	startTime := time.Now()

	// 70人の参加者を同時に登録
	for i := 0; i < numParticipants; i++ {
		wg.Add(1)
		go func(participantNum int) {
			defer wg.Done()

			reqStart := time.Now()

			// 参加者登録リクエスト
			participantReq := models.ParticipantRequest{
				Nickname: fmt.Sprintf("LoadTestUser%d", participantNum),
			}

			jsonData, err := json.Marshal(participantReq)
			if err != nil {
				results <- RequestResult{
					Success:   false,
					Latency:   time.Since(reqStart),
					Error:     err,
					Timestamp: time.Now(),
				}
				return
			}

			client := createHTTPClient()
			resp, err := client.Post(
				BaseURL+"/api/participants/register",
				"application/json",
				bytes.NewBuffer(jsonData),
			)

			latency := time.Since(reqStart)
			success := err == nil && resp != nil && resp.StatusCode == http.StatusCreated

			if resp != nil {
				resp.Body.Close()
			}

			results <- RequestResult{
				Success:   success,
				Latency:   latency,
				Error:     err,
				Timestamp: time.Now(),
			}
		}(i)
	}

	// 全ての goroutine の完了を待つ
	wg.Wait()
	close(results)

	// 結果を集計
	var totalLatency time.Duration
	var maxLatency time.Duration
	minLatency := time.Hour // 初期値として大きな値を設定
	successCount := 0
	failCount := 0

	for result := range results {
		if result.Success {
			successCount++
		} else {
			failCount++
			if result.Error != nil {
				t.Logf("Request failed: %v", result.Error)
			}
		}

		totalLatency += result.Latency
		if result.Latency > maxLatency {
			maxLatency = result.Latency
		}
		if result.Latency < minLatency {
			minLatency = result.Latency
		}
	}

	totalDuration := time.Since(startTime)
	avgLatency := totalLatency / time.Duration(numParticipants)
	requestsPerSec := float64(numParticipants) / totalDuration.Seconds()
	errorRate := float64(failCount) / float64(numParticipants) * 100

	// 結果を出力
	t.Logf("=== Concurrent Participant Registration Test Results ===")
	t.Logf("Total Participants: %d", numParticipants)
	t.Logf("Successful Registrations: %d", successCount)
	t.Logf("Failed Registrations: %d", failCount)
	t.Logf("Total Duration: %v", totalDuration)
	t.Logf("Average Latency: %v", avgLatency)
	t.Logf("Max Latency: %v", maxLatency)
	t.Logf("Min Latency: %v", minLatency)
	t.Logf("Requests per Second: %.2f", requestsPerSec)
	t.Logf("Error Rate: %.2f%%", errorRate)

	// パフォーマンス基準をチェック
	if errorRate > 5.0 {
		t.Errorf("Error rate too high: %.2f%% (expected < 5%%)", errorRate)
	}

	if avgLatency > 5*time.Second {
		t.Errorf("Average latency too high: %v (expected < 5s)", avgLatency)
	}

	if successCount < int(float64(numParticipants)*0.95) {
		t.Errorf("Success rate too low: %d/%d (expected > 95%%)", successCount, numParticipants)
	}
}

func TestConcurrentWebSocketConnections(t *testing.T) {
	if testing.Short() && os.Getenv("RUN_PERFORMANCE_TESTS") != runPerfTestsEnv {
		t.Skip("Skipping performance test in short mode")
	}

	setupPerformanceTest(t)
	defer cleanupPerformanceTest(t)

	numConnections := getMaxConcurrentUsers()
	connections := make([]*websocket.Conn, 0, numConnections)
	var wg sync.WaitGroup
	var mu sync.Mutex
	connectSuccessCount := 0
	connectFailCount := 0

	startTime := time.Now()

	// 70個のWebSocket接続を同時に確立
	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(connNum int) {
			defer wg.Done()

			u, err := url.Parse(WebSocketURL)
			if err != nil {
				mu.Lock()
				connectFailCount++
				mu.Unlock()
				t.Logf("URL parse error: %v", err)
				return
			}

			dialer := createWebSocketDialer()
			conn, resp, err := dialer.Dial(u.String(), nil)
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			if err != nil {
				mu.Lock()
				connectFailCount++
				mu.Unlock()
				t.Logf("WebSocket connection failed for conn %d: %v", connNum, err)
				return
			}

			mu.Lock()
			connections = append(connections, conn)
			connectSuccessCount++
			mu.Unlock()

			// 接続後にハートビートメッセージを送信
			heartbeat := map[string]interface{}{
				"type": "heartbeat",
			}

			if err := conn.WriteJSON(heartbeat); err != nil {
				t.Logf("Failed to send heartbeat for conn %d: %v", connNum, err)
			}

			// 短時間接続を維持
			time.Sleep(2 * time.Second)
		}(i)
	}

	wg.Wait()
	connectionDuration := time.Since(startTime)

	t.Logf("=== Concurrent WebSocket Connection Test Results ===")
	t.Logf("Target Connections: %d", numConnections)
	t.Logf("Successful Connections: %d", connectSuccessCount)
	t.Logf("Failed Connections: %d", connectFailCount)
	t.Logf("Connection Duration: %v", connectionDuration)
	t.Logf("Connection Success Rate: %.2f%%", float64(connectSuccessCount)/float64(numConnections)*100)

	// WebSocketメッセージ送信テスト
	if len(connections) > 0 {
		messagesSent := 0
		messagesSuccess := 0

		// 各接続でクイズ購読メッセージを送信
		for i, conn := range connections {
			subscribeMsg := map[string]interface{}{
				"type":    "subscribe",
				"quiz_id": 2,
			}

			messagesSent++
			if err := conn.WriteJSON(subscribeMsg); err != nil {
				t.Logf("Failed to send subscribe message for conn %d: %v", i, err)
			} else {
				messagesSuccess++
			}
		}

		t.Logf("Messages Sent: %d", messagesSent)
		t.Logf("Messages Success: %d", messagesSuccess)
		t.Logf("Message Success Rate: %.2f%%", float64(messagesSuccess)/float64(messagesSent)*100)
	}

	// 接続をクリーンアップ
	for i, conn := range connections {
		if err := conn.Close(); err != nil {
			t.Logf("Failed to close connection %d: %v", i, err)
		}
	}

	// パフォーマンス基準をチェック
	connectionSuccessRate := float64(connectSuccessCount) / float64(numConnections) * 100
	if connectionSuccessRate < 95.0 {
		t.Errorf("WebSocket connection success rate too low: %.2f%% (expected > 95%%)", connectionSuccessRate)
	}

	if connectionDuration > 10*time.Second {
		t.Errorf("Connection establishment took too long: %v (expected < 10s)", connectionDuration)
	}
}

//nolint:gocyclo
func TestConcurrentAnswerSubmissions(t *testing.T) {
	if testing.Short() && os.Getenv("RUN_PERFORMANCE_TESTS") != runPerfTestsEnv {
		t.Skip("Skipping performance test in short mode")
	}

	setupPerformanceTest(t)
	defer cleanupPerformanceTest(t)

	numParticipants := getMaxConcurrentUsers()
	var participantIDs []int64
	var wg sync.WaitGroup

	// まず参加者を登録
	t.Log("Registering participants for answer submission test...")
	client := createHTTPClient()
	for i := 0; i < numParticipants; i++ {
		participantReq := models.ParticipantRequest{
			Nickname: fmt.Sprintf("AnswerTestUser%d", i),
		}

		jsonData, _ := json.Marshal(participantReq)
		resp, err := client.Post(
			BaseURL+"/api/participants/register",
			"application/json",
			bytes.NewBuffer(jsonData),
		)

		if err == nil && resp.StatusCode == http.StatusCreated {
			var result map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&result) // テスト用なのでエラーハンドリング不要

			if data, ok := result["data"].(map[string]interface{}); ok {
				if participantID, ok := data["participant_id"].(float64); ok {
					participantIDs = append(participantIDs, int64(participantID))
				}
			}
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if len(participantIDs) == 0 {
		t.Fatal("No participants registered for answer submission test")
	}

	t.Logf("Registered %d participants for answer submission test", len(participantIDs))

	// 同時回答送信テスト
	results := make(chan RequestResult, len(participantIDs))
	startTime := time.Now()

	for i, participantID := range participantIDs {
		wg.Add(1)
		go func(pID int64, userNum int) {
			defer wg.Done()

			reqStart := time.Now()

			answerReq := models.AnswerRequest{
				ParticipantID:  pID,
				QuizID:         2,
				SelectedOption: []string{"A", "B", "C", "D"}[userNum%4],
			}

			jsonData, err := json.Marshal(answerReq)
			if err != nil {
				results <- RequestResult{
					Success:   false,
					Latency:   time.Since(reqStart),
					Error:     err,
					Timestamp: time.Now(),
				}
				return
			}

			client := createHTTPClient()
			resp, err := client.Post(
				BaseURL+"/api/answers",
				"application/json",
				bytes.NewBuffer(jsonData),
			)

			latency := time.Since(reqStart)
			success := err == nil && resp != nil && (resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK)

			if resp != nil {
				resp.Body.Close()
			}

			results <- RequestResult{
				Success:   success,
				Latency:   latency,
				Error:     err,
				Timestamp: time.Now(),
			}
		}(participantID, i)
	}

	wg.Wait()
	close(results)

	// 結果集計
	var totalLatency time.Duration
	var maxLatency time.Duration
	minLatency := time.Hour
	successCount := 0
	failCount := 0

	for result := range results {
		if result.Success {
			successCount++
		} else {
			failCount++
			if result.Error != nil {
				t.Logf("Answer submission failed: %v", result.Error)
			}
		}

		totalLatency += result.Latency
		if result.Latency > maxLatency {
			maxLatency = result.Latency
		}
		if result.Latency < minLatency {
			minLatency = result.Latency
		}
	}

	totalDuration := time.Since(startTime)
	avgLatency := totalLatency / time.Duration(len(participantIDs))
	requestsPerSec := float64(len(participantIDs)) / totalDuration.Seconds()
	errorRate := float64(failCount) / float64(len(participantIDs)) * 100

	t.Logf("=== Concurrent Answer Submission Test Results ===")
	t.Logf("Total Answers: %d", len(participantIDs))
	t.Logf("Successful Submissions: %d", successCount)
	t.Logf("Failed Submissions: %d", failCount)
	t.Logf("Total Duration: %v", totalDuration)
	t.Logf("Average Latency: %v", avgLatency)
	t.Logf("Max Latency: %v", maxLatency)
	t.Logf("Min Latency: %v", minLatency)
	t.Logf("Requests per Second: %.2f", requestsPerSec)
	t.Logf("Error Rate: %.2f%%", errorRate)

	// パフォーマンス基準をチェック
	if errorRate > 5.0 {
		t.Errorf("Answer submission error rate too high: %.2f%% (expected < 5%%)", errorRate)
	}

	if avgLatency > 3*time.Second {
		t.Errorf("Answer submission average latency too high: %v (expected < 3s)", avgLatency)
	}
}

//nolint:gocyclo
func TestSystemLoadUnder70Users(t *testing.T) {
	if testing.Short() && os.Getenv("RUN_PERFORMANCE_TESTS") != runPerfTestsEnv {
		t.Skip("Skipping performance test in short mode")
	}

	setupPerformanceTest(t)
	defer cleanupPerformanceTest(t)

	t.Log("=== System Load Test Under Concurrent Users ===")

	// システム全体の負荷テスト（GitHub Actions環境を考慮して軽量化）
	numUsers := getMaxConcurrentUsers()
	
	// GitHub Actions環境ではより軽量な設定
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		numUsers = 20  // 50から20に削減
		t.Logf("GitHub Actions環境: ユーザー数を %d に削減", numUsers)
	}

	var wg sync.WaitGroup
	results := make(chan RequestResult, numUsers*5) // バッファサイズも削減
	startTime := time.Now()

	t.Logf("開始: %d人の同時負荷テスト", numUsers)

	// 各ユーザーが以下のアクションを実行:
	// 1. 参加者登録
	// 2. 制限された回答送信
	// 3. セッション状況確認
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userNum int) {
			defer wg.Done()

			userNickname := fmt.Sprintf("LoadTestUser%d", userNum)
			client := createHTTPClient()

			t.Logf("ユーザー%d: 開始", userNum)

			// 1. 参加者登録
			participantReq := models.ParticipantRequest{Nickname: userNickname}
			jsonData, _ := json.Marshal(participantReq)

			reqStart := time.Now()
			resp, err := client.Post(BaseURL+"/api/participants/register", "application/json", bytes.NewBuffer(jsonData))

			results <- RequestResult{
				Success:   err == nil && resp != nil && resp.StatusCode == http.StatusCreated,
				Latency:   time.Since(reqStart),
				Error:     err,
				Timestamp: time.Now(),
			}

			var participantID int64
			if resp != nil && resp.StatusCode == http.StatusCreated {
				var result map[string]interface{}
				_ = json.NewDecoder(resp.Body).Decode(&result) // テスト用なのでエラーハンドリング不要
				if data, ok := result["data"].(map[string]interface{}); ok {
					if pID, ok := data["participant_id"].(float64); ok {
						participantID = int64(pID)
					}
				}
				resp.Body.Close()
			}

			// 2. 制限された回答送信（WebSocket接続を避けて軽量化）
			if participantID > 0 {
				// 回答送信テスト（最大3回）
				maxAnswers := 3
				for i := 0; i < maxAnswers; i++ {
					answerReq := models.AnswerRequest{
						ParticipantID:  participantID,
						QuizID:         2,
						SelectedOption: []string{"A", "B", "C", "D"}[userNum%4],
					}

					jsonData, _ := json.Marshal(answerReq)
					reqStart := time.Now()
					resp, err := client.Post(BaseURL+"/api/answers", "application/json", bytes.NewBuffer(jsonData))

					results <- RequestResult{
						Success:   err == nil && resp != nil && (resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK),
						Latency:   time.Since(reqStart),
						Error:     err,
						Timestamp: time.Now(),
					}

					if resp != nil {
						resp.Body.Close()
					}

					time.Sleep(200 * time.Millisecond) // 間隔調整
				}
			}

			// 3. セッション状況確認（最大2回）
			for i := 0; i < 2; i++ {
				reqStart := time.Now()
				resp, err := client.Get(BaseURL + "/api/session/status")

				results <- RequestResult{
					Success:   err == nil && resp != nil && resp.StatusCode == http.StatusOK,
					Latency:   time.Since(reqStart),
					Error:     err,
					Timestamp: time.Now(),
				}

				if resp != nil {
					resp.Body.Close()
				}

				time.Sleep(500 * time.Millisecond)
			}

			t.Logf("ユーザー%d: 完了", userNum)
		}(i)
	}

	wg.Wait()
	close(results)

	totalDuration := time.Since(startTime)

	// 結果集計
	var totalLatency time.Duration
	totalRequests := 0
	successCount := 0
	failCount := 0
	var maxLatency time.Duration

	for result := range results {
		totalRequests++
		totalLatency += result.Latency

		if result.Success {
			successCount++
		} else {
			failCount++
		}

		if result.Latency > maxLatency {
			maxLatency = result.Latency
		}
	}

	avgLatency := totalLatency / time.Duration(totalRequests)
	requestsPerSec := float64(totalRequests) / totalDuration.Seconds()
	errorRate := float64(failCount) / float64(totalRequests) * 100

	t.Logf("=== System Load Test Results ===")
	t.Logf("Test Duration: %v", totalDuration)
	t.Logf("Concurrent Users: %d", numUsers)
	t.Logf("Total Requests: %d", totalRequests)
	t.Logf("Successful Requests: %d", successCount)
	t.Logf("Failed Requests: %d", failCount)
	t.Logf("Average Latency: %v", avgLatency)
	t.Logf("Max Latency: %v", maxLatency)
	t.Logf("Requests per Second: %.2f", requestsPerSec)
	t.Logf("Error Rate: %.2f%%", errorRate)
	t.Logf("Throughput: %.2f req/user/sec", requestsPerSec/float64(numUsers))

	// システム負荷の基準チェック
	if errorRate > 2.0 {
		t.Errorf("System error rate under load too high: %.2f%% (expected < 2%%)", errorRate)
	}

	if avgLatency > 2*time.Second {
		t.Errorf("System average latency under load too high: %v (expected < 2s)", avgLatency)
	}

	if requestsPerSec < 50.0 {
		t.Errorf("System throughput too low: %.2f req/sec (expected > 50 req/sec)", requestsPerSec)
	}
}
