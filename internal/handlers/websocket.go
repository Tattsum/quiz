package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// MaxConnections is the maximum number of concurrent WebSocket connections allowed
	MaxConnections = 70 // 最大接続数
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(_ /*r*/ *http.Request) bool {
			// In production, implement proper origin checking
			return true
		},
	}

	// Store active WebSocket connections
	connections      = make(map[*websocket.Conn]*ClientConnection)
	connectionsMutex = sync.RWMutex{}
)

// ClientConnection represents a WebSocket client connection
type ClientConnection struct {
	Conn          *websocket.Conn
	QuizID        *int64
	LastHeartbeat time.Time
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// SubscribeMessage represents a subscription message
type SubscribeMessage struct {
	Type   string `json:"type"`
	QuizID int64  `json:"quiz_id"`
}

// QuestionSwitchNotification represents a question switch notification
type QuestionSwitchNotification struct {
	QuizID         int64     `json:"quiz_id"`
	QuestionNumber int       `json:"question_number"`
	TotalQuestions int       `json:"total_questions"`
	SwitchedAt     time.Time `json:"switched_at"`
}

// VotingEndNotification represents a voting end notification
type VotingEndNotification struct {
	QuizID     int64     `json:"quiz_id"`
	QuestionID int64     `json:"question_id"`
	EndedAt    time.Time `json:"ended_at"`
}

// AnswerStatusUpdate represents current answer status
type AnswerStatusUpdate struct {
	QuizID            int64          `json:"quiz_id"`
	QuestionID        int64          `json:"question_id"`
	TotalParticipants int            `json:"total_participants"`
	AnsweredCount     int            `json:"answered_count"`
	AnswerCounts      map[string]int `json:"answer_counts"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// WebSocketResults handles WebSocket connections for real-time results
func WebSocketResults(c *gin.Context) {
	// Check connection limit
	connectionsMutex.RLock()
	currentConnections := len(connections)
	connectionsMutex.RUnlock()

	if currentConnections >= MaxConnections {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":           "Maximum connections reached",
			"max_connections": MaxConnections,
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close WebSocket connection: %v", err)
		}
	}()

	// Register connection
	client := &ClientConnection{
		Conn:          conn,
		LastHeartbeat: time.Now(),
	}

	connectionsMutex.Lock()
	connections[conn] = client
	log.Printf("New WebSocket connection established. Total connections: %d/%d", len(connections), MaxConnections)
	connectionsMutex.Unlock()

	// Remove connection when done
	defer func() {
		connectionsMutex.Lock()
		delete(connections, conn)
		log.Printf("WebSocket connection closed. Total connections: %d/%d", len(connections), MaxConnections)
		connectionsMutex.Unlock()
	}()

	// Set up ping/pong for connection health check
	conn.SetPongHandler(func(string) error {
		client.LastHeartbeat = time.Now()
		return nil
	})

	// Start heartbeat goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// Handle incoming messages
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			var msg SubscribeMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			switch msg.Type {
			case "subscribe":
				client.QuizID = &msg.QuizID

				// Send current results immediately
				results, err := getCurrentQuizResults(msg.QuizID)
				if err == nil {
					sendMessage(conn, "result_update", results)
				}

			case "unsubscribe":
				client.QuizID = nil

			case "heartbeat":
				client.LastHeartbeat = time.Now()
				sendMessage(conn, "heartbeat_ack", map[string]interface{}{
					"timestamp": time.Now(),
				})
			}
		}
	}
}

// BroadcastResultUpdate broadcasts result updates to all subscribed clients
func BroadcastResultUpdate(quizID int64) {
	results, err := getCurrentQuizResults(quizID)
	if err != nil {
		log.Printf("Failed to get quiz results for broadcast: %v", err)
		return
	}

	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()

	for conn, client := range connections {
		if client.QuizID != nil && *client.QuizID == quizID {
			go func(c *websocket.Conn) {
				sendMessage(c, "result_update", results)
			}(conn)
		}
	}
}

// BroadcastSessionUpdate broadcasts session status updates to all clients
func BroadcastSessionUpdate(sessionData interface{}) {
	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()

	for conn := range connections {
		go func(c *websocket.Conn) {
			sendMessage(c, "session_update", sessionData)
		}(conn)
	}
}

// BroadcastQuestionSwitch broadcasts question switch notifications
func BroadcastQuestionSwitch(quizID int64, questionNumber, totalQuestions int) {
	notification := QuestionSwitchNotification{
		QuizID:         quizID,
		QuestionNumber: questionNumber,
		TotalQuestions: totalQuestions,
		SwitchedAt:     time.Now(),
	}

	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()

	for conn, client := range connections {
		if client.QuizID != nil && *client.QuizID == quizID {
			go func(c *websocket.Conn) {
				sendMessage(c, "question_switch", notification)
			}(conn)
		}
	}

	log.Printf("Broadcasted question switch for quiz %d to %d subscribers", quizID, GetSubscriptionCount(quizID))
}

// BroadcastVotingEnd broadcasts voting end notifications
func BroadcastVotingEnd(quizID, questionID int64) {
	notification := VotingEndNotification{
		QuizID:     quizID,
		QuestionID: questionID,
		EndedAt:    time.Now(),
	}

	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()

	for conn, client := range connections {
		if client.QuizID != nil && *client.QuizID == quizID {
			go func(c *websocket.Conn) {
				sendMessage(c, "voting_end", notification)
			}(conn)
		}
	}

	log.Printf("Broadcasted voting end for quiz %d, question %d to %d subscribers", quizID, questionID, GetSubscriptionCount(quizID))
}

// BroadcastAnswerStatus broadcasts current answer status
func BroadcastAnswerStatus(quizID, questionID int64, totalParticipants, answeredCount int, answerCounts map[string]int) {
	status := AnswerStatusUpdate{
		QuizID:            quizID,
		QuestionID:        questionID,
		TotalParticipants: totalParticipants,
		AnsweredCount:     answeredCount,
		AnswerCounts:      answerCounts,
		UpdatedAt:         time.Now(),
	}

	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()

	for conn, client := range connections {
		if client.QuizID != nil && *client.QuizID == quizID {
			go func(c *websocket.Conn) {
				sendMessage(c, "answer_status", status)
			}(conn)
		}
	}
}

// sendMessage sends a message to a WebSocket connection
func sendMessage(conn *websocket.Conn, messageType string, data interface{}) {
	message := WebSocketMessage{
		Type: messageType,
		Data: data,
	}

	if err := conn.WriteJSON(message); err != nil {
		log.Printf("Failed to send WebSocket message: %v", err)
	}
}

// getCurrentQuizResults gets current results for a quiz
func getCurrentQuizResults(quizID int64) (*models.QuizResultsResponse, error) {
	db := database.GetDB()
	return getQuizResultsData(db, quizID, nil)
}

// CleanupConnections removes stale WebSocket connections
func CleanupConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			connectionsMutex.Lock()
			cutoff := time.Now().Add(-2 * time.Minute)

			for conn, client := range connections {
				if client.LastHeartbeat.Before(cutoff) {
					_ = conn.Close()
					delete(connections, conn)
				}
			}
			connectionsMutex.Unlock()
		}
	}
}

// GetConnectionCount returns the current number of active WebSocket connections
func GetConnectionCount() int {
	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()
	return len(connections)
}

// GetSubscriptionCount returns the number of connections subscribed to a specific quiz
func GetSubscriptionCount(quizID int64) int {
	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()

	count := 0
	for _, client := range connections {
		if client.QuizID != nil && *client.QuizID == quizID {
			count++
		}
	}
	return count
}

// init initializes the WebSocket cleanup goroutine
func init() {
	go CleanupConnections()
}
