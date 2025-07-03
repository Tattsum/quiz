package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// In production, implement proper origin checking
			return true
		},
	}
	
	// Store active WebSocket connections
	connections = make(map[*websocket.Conn]*ClientConnection)
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

// WebSocketResults handles WebSocket connections for real-time results
func WebSocketResults(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Register connection
	client := &ClientConnection{
		Conn:          conn,
		LastHeartbeat: time.Now(),
	}
	
	connectionsMutex.Lock()
	connections[conn] = client
	connectionsMutex.Unlock()

	// Remove connection when done
	defer func() {
		connectionsMutex.Lock()
		delete(connections, conn)
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
					conn.Close()
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