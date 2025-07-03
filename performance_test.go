package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/Tattsum/quiz/internal/models"
	"github.com/gorilla/websocket"
)

const (
	// パフォーマンステスト用の設定
	MaxConcurrentUsers = 70
	BaseURL            = "http://localhost:8080"
	WebSocketURL       = "ws://localhost:8080/ws"
	TestDuration       = 30 * time.Second
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

func TestConcurrentParticipantRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	numParticipants := MaxConcurrentUsers
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

			resp, err := http.Post(
				BaseURL+"/api/public/participants",
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
	var minLatency time.Duration = time.Hour // 初期値として大きな値を設定
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
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	numConnections := MaxConcurrentUsers
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

			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
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
				"quiz_id": 1,
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

func TestConcurrentAnswerSubmissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	numParticipants := MaxConcurrentUsers
	var participantIDs []int64
	var wg sync.WaitGroup

	// まず参加者を登録
	t.Log("Registering participants for answer submission test...")
	for i := 0; i < numParticipants; i++ {
		participantReq := models.ParticipantRequest{
			Nickname: fmt.Sprintf("AnswerTestUser%d", i),
		}

		jsonData, _ := json.Marshal(participantReq)
		resp, err := http.Post(
			BaseURL+"/api/public/participants",
			"application/json",
			bytes.NewBuffer(jsonData),
		)

		if err == nil && resp.StatusCode == http.StatusCreated {
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)

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
				QuizID:         1,
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

			resp, err := http.Post(
				BaseURL+"/api/public/answers",
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
	var minLatency time.Duration = time.Hour
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

func TestSystemLoadUnder70Users(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Log("=== System Load Test Under 70 Concurrent Users ===")

	// システム全体の負荷テスト
	duration := TestDuration
	numUsers := MaxConcurrentUsers

	var wg sync.WaitGroup
	results := make(chan RequestResult, numUsers*10) // 各ユーザーが複数リクエストを送信
	startTime := time.Now()

	// 各ユーザーが以下のアクションを実行:
	// 1. 参加者登録
	// 2. WebSocket接続
	// 3. 複数回の回答送信
	// 4. データ取得リクエスト
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userNum int) {
			defer wg.Done()

			userStartTime := time.Now()
			userNickname := fmt.Sprintf("LoadTestUser%d", userNum)

			// 1. 参加者登録
			participantReq := models.ParticipantRequest{Nickname: userNickname}
			jsonData, _ := json.Marshal(participantReq)

			reqStart := time.Now()
			resp, err := http.Post(BaseURL+"/api/public/participants", "application/json", bytes.NewBuffer(jsonData))

			results <- RequestResult{
				Success:   err == nil && resp != nil && resp.StatusCode == http.StatusCreated,
				Latency:   time.Since(reqStart),
				Error:     err,
				Timestamp: time.Now(),
			}

			var participantID int64
			if resp != nil && resp.StatusCode == http.StatusCreated {
				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				if data, ok := result["data"].(map[string]interface{}); ok {
					if pID, ok := data["participant_id"].(float64); ok {
						participantID = int64(pID)
					}
				}
				resp.Body.Close()
			}

			// 2. WebSocket接続
			if participantID > 0 {
				u, _ := url.Parse(WebSocketURL)
				conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

				if err == nil && conn != nil {
					defer conn.Close()

					// ハートビート送信
					heartbeat := map[string]interface{}{"type": "heartbeat"}
					conn.WriteJSON(heartbeat)

					// 3. 複数回の回答送信（時間内で）
					for time.Since(userStartTime) < duration/2 {
						answerReq := models.AnswerRequest{
							ParticipantID:  participantID,
							QuizID:         1,
							SelectedOption: []string{"A", "B", "C", "D"}[userNum%4],
						}

						jsonData, _ := json.Marshal(answerReq)
						reqStart := time.Now()
						resp, err := http.Post(BaseURL+"/api/public/answers", "application/json", bytes.NewBuffer(jsonData))

						results <- RequestResult{
							Success:   err == nil && resp != nil && (resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK),
							Latency:   time.Since(reqStart),
							Error:     err,
							Timestamp: time.Now(),
						}

						if resp != nil {
							resp.Body.Close()
						}

						time.Sleep(100 * time.Millisecond) // リクエスト間隔
					}
				}
			}

			// 4. データ取得リクエスト
			for time.Since(userStartTime) < duration {
				reqStart := time.Now()
				resp, err := http.Get(BaseURL + "/api/public/session/status")

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
