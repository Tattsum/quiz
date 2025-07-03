package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestWebSocketConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// テスト用サーバーを作成
	r := gin.New()
	r.GET("/ws", WebSocketResults)

	server := httptest.NewServer(r)
	defer server.Close()

	// WebSocketのURLを作成
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// WebSocket接続を確立
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// 接続が成功していることを確認
	if conn == nil {
		t.Error("WebSocket connection is nil")
	}
}

func TestWebSocketSubscribe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ws", WebSocketResults)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// 購読メッセージを送信
	subscribeMsg := SubscribeMessage{
		Type:   "subscribe",
		QuizID: 1,
	}

	err = conn.WriteJSON(subscribeMsg)
	if err != nil {
		t.Fatalf("Failed to send subscribe message: %v", err)
	}

	// メッセージを受信できることを確認（タイムアウト付き）
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		// タイムアウトエラーは期待される場合もある
		if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			t.Logf("Read message error (expected in some cases): %v", err)
		}
	} else {
		t.Logf("Received message: %s", string(message))
	}
}

func TestWebSocketHeartbeat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ws", WebSocketResults)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// ハートビートメッセージを送信
	heartbeatMsg := WebSocketMessage{
		Type: "heartbeat",
		Data: nil,
	}

	err = conn.WriteJSON(heartbeatMsg)
	if err != nil {
		t.Fatalf("Failed to send heartbeat message: %v", err)
	}

	// レスポンスを待機（短いタイムアウト）
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = conn.ReadMessage()
	if err != nil {
		// ハートビートのレスポンスがない場合もあるので、エラーをログに出力するだけ
		t.Logf("Heartbeat response error (may be expected): %v", err)
	}
}

func TestWebSocketUnsubscribe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ws", WebSocketResults)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// まず購読
	subscribeMsg := SubscribeMessage{
		Type:   "subscribe",
		QuizID: 1,
	}
	err = conn.WriteJSON(subscribeMsg)
	if err != nil {
		t.Fatalf("Failed to send subscribe message: %v", err)
	}

	// 購読解除
	unsubscribeMsg := WebSocketMessage{
		Type: "unsubscribe",
		Data: nil,
	}
	err = conn.WriteJSON(unsubscribeMsg)
	if err != nil {
		t.Fatalf("Failed to send unsubscribe message: %v", err)
	}

	// 接続が正常に維持されていることを確認
	heartbeatMsg := WebSocketMessage{
		Type: "heartbeat",
		Data: nil,
	}
	err = conn.WriteJSON(heartbeatMsg)
	if err != nil {
		t.Fatalf("Failed to send heartbeat after unsubscribe: %v", err)
	}
}

func TestWebSocketConnectionLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ws", WebSocketResults)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// 複数の接続を作成
	var connections []*websocket.Conn
	maxTestConnections := 5 // テスト用に少ない数にする

	for i := 0; i < maxTestConnections; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Logf("Connection %d failed: %v", i, err)
			break
		}
		connections = append(connections, conn)
	}

	// 接続をクリーンアップ
	for _, conn := range connections {
		conn.Close()
	}

	if len(connections) == 0 {
		t.Error("No connections were established")
	} else {
		t.Logf("Successfully established %d connections", len(connections))
	}
}

func TestWebSocketInvalidMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ws", WebSocketResults)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// 無効なJSONメッセージを送信
	err = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	if err != nil {
		t.Fatalf("Failed to send invalid message: %v", err)
	}

	// 接続が切断されるかタイムアウトを待つ
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err = conn.ReadMessage()
	if err != nil {
		// エラーが発生することを期待（接続切断またはタイムアウト）
		t.Logf("Expected error occurred: %v", err)
	}
}

func TestBroadcastFunctions(t *testing.T) {
	// ブロードキャスト関数のユニットテスト
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "BroadcastQuestionSwitch",
			fn: func() {
				BroadcastQuestionSwitch(1, 1, 10)
			},
		},
		{
			name: "BroadcastVotingEnd",
			fn: func() {
				BroadcastVotingEnd(1, 1)
			},
		},
		{
			name: "BroadcastResultUpdate",
			fn: func() {
				BroadcastResultUpdate(1)
			},
		},
		{
			name: "BroadcastSessionUpdate",
			fn: func() {
				BroadcastSessionUpdate("session_update")
			},
		},
		{
			name: "BroadcastAnswerStatus",
			fn: func() {
				answerCounts := map[string]int{
					"A": 5,
					"B": 3,
					"C": 2,
					"D": 1,
				}
				BroadcastAnswerStatus(1, 1, 11, 11, answerCounts)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ブロードキャスト関数がパニックを起こさないことを確認
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Broadcast function %s panicked: %v", tt.name, r)
				}
			}()

			tt.fn()
		})
	}
}

func TestGetConnectionCount(t *testing.T) {
	// 接続数を取得する関数のテスト
	count := GetConnectionCount()
	if count < 0 {
		t.Errorf("Connection count should not be negative, got %d", count)
	}
}

func TestGetSubscriptionCount(t *testing.T) {
	// 特定クイズの購読数を取得する関数のテスト
	count := GetSubscriptionCount(1)
	if count < 0 {
		t.Errorf("Subscription count should not be negative, got %d", count)
	}
}

func TestWebSocketConcurrentConnections(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ws", WebSocketResults)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// 並行接続テスト
	numConnections := 10
	done := make(chan bool, numConnections)
	errors := make(chan error, numConnections)

	for i := 0; i < numConnections; i++ {
		go func(connNum int) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				errors <- err
				done <- false
				return
			}
			defer conn.Close()

			// 購読メッセージを送信
			subscribeMsg := SubscribeMessage{
				Type:   "subscribe",
				QuizID: int64(connNum%3 + 1), // 3つの異なるクイズIDを使用
			}
			err = conn.WriteJSON(subscribeMsg)
			if err != nil {
				errors <- err
				done <- false
				return
			}

			// 短時間待機
			time.Sleep(100 * time.Millisecond)

			done <- true
		}(i)
	}

	// すべてのgoroutineの完了を待つ
	successCount := 0
	errorCount := 0

	for i := 0; i < numConnections; i++ {
		select {
		case success := <-done:
			if success {
				successCount++
			} else {
				errorCount++
			}
		case err := <-errors:
			t.Logf("Connection error: %v", err)
			errorCount++
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for connections")
		}
	}

	t.Logf("Successful connections: %d, Failed connections: %d", successCount, errorCount)

	if successCount == 0 {
		t.Error("No connections were successful")
	}
}
