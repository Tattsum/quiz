package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
	"github.com/gin-gonic/gin"
)

func TestRegisterParticipant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// テスト環境用のデータベース設定
	setupTestEnv()

	// データベース接続を初期化
	_, err := database.Initialize()
	if err != nil && os.Getenv("TEST_ENV") != testEnvValue {
		t.Skipf("Database connection failed (not in test environment): %v", err)
	} else if err != nil {
		t.Fatalf("Database connection failed in test environment: %v", err)
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "Register participant with valid nickname",
			requestBody: models.ParticipantRequest{
				Nickname: "TestUser",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Register participant with empty nickname",
			requestBody: models.ParticipantRequest{
				Nickname: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Register participant with too long nickname",
			requestBody: models.ParticipantRequest{
				Nickname: "ThisNicknameIsWayTooLongAndShouldFailValidation12345",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Register participant with malformed JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req, _ := http.NewRequest("POST", "/participants", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			RegisterParticipant(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestGetParticipant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// テスト環境用のデータベース設定
	setupTestEnv()

	// データベース接続を初期化
	_, err := database.Initialize()
	if err != nil && os.Getenv("TEST_ENV") != testEnvValue {
		t.Skipf("Database connection failed (not in test environment): %v", err)
	} else if err != nil {
		t.Fatalf("Database connection failed in test environment: %v", err)
	}

	// テスト用参加者を作成
	db := database.GetDB()
	_, err = db.Exec(`
		INSERT INTO participants (id, nickname, created_at)
		VALUES (1, 'TestUser1', CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			nickname = 'TestUser1',
			created_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		t.Fatalf("Failed to create test participant: %v", err)
	}

	tests := []struct {
		name           string
		participantID  string
		expectedStatus int
	}{
		{
			name:           "Get participant with valid ID",
			participantID:  "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get participant with invalid ID",
			participantID:  "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get participant with non-existent ID",
			participantID:  "999999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{
				{Key: "id", Value: tt.participantID},
			}

			GetParticipant(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestGetParticipantAnswers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// テスト環境用のデータベース設定
	setupTestEnv()

	// データベース接続を初期化
	_, err := database.Initialize()
	if err != nil && os.Getenv("TEST_ENV") != testEnvValue {
		t.Skipf("Database connection failed (not in test environment): %v", err)
	} else if err != nil {
		t.Fatalf("Database connection failed in test environment: %v", err)
	}

	// テスト用参加者を作成
	db := database.GetDB()
	_, err = db.Exec(`
		INSERT INTO participants (id, nickname, created_at)
		VALUES (1, 'TestUser1', CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			nickname = 'TestUser1',
			created_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		t.Fatalf("Failed to create test participant: %v", err)
	}

	tests := []struct {
		name           string
		participantID  string
		expectedStatus int
	}{
		{
			name:           "Get answers for valid participant ID",
			participantID:  "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get answers for invalid participant ID",
			participantID:  "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get answers for non-existent participant ID",
			participantID:  "999999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{
				{Key: "id", Value: tt.participantID},
			}

			GetParticipantAnswers(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestSubmitAnswer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// テスト環境用のデータベース設定
	setupTestEnv()

	// データベース接続を初期化
	_, err := database.Initialize()
	if err != nil && os.Getenv("TEST_ENV") != testEnvValue {
		t.Skipf("Database connection failed (not in test environment): %v", err)
	} else if err != nil {
		t.Fatalf("Database connection failed in test environment: %v", err)
	}

	// テスト用データを確実に作成
	db := database.GetDB()

	// テスト用参加者を作成
	_, err = db.Exec(`
		INSERT INTO participants (id, nickname, created_at)
		VALUES (1, 'TestUser1', CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			nickname = 'TestUser1',
			created_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		t.Fatalf("Failed to create test participant: %v", err)
	}

	// テスト用クイズを作成
	_, err = db.Exec(`
		INSERT INTO quizzes (id, question_text, option_a, option_b, option_c, option_d, correct_answer, created_at, updated_at)
		VALUES (1, 'Test Question?', 'Option A', 'Option B', 'Option C', 'Option D', 'A', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			question_text = 'Test Question?',
			option_a = 'Option A',
			option_b = 'Option B',
			option_c = 'Option C',
			option_d = 'Option D',
			correct_answer = 'A',
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		t.Fatalf("Failed to create test quiz: %v", err)
	}

	// テスト用のセッションを開始（既存のセッションと回答を削除してから作成）
	_, err = db.Exec(`DELETE FROM answers`)
	if err != nil {
		t.Fatalf("Failed to clear test answers: %v", err)
	}
	_, err = db.Exec(`DELETE FROM quiz_sessions`)
	if err != nil {
		t.Fatalf("Failed to clear test sessions: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO quiz_sessions (current_quiz_id, is_accepting_answers, created_at, updated_at)
		VALUES (1, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		t.Fatalf("Failed to start test session: %v", err)
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "Submit answer with valid data",
			requestBody: models.AnswerRequest{
				ParticipantID:  1, // 既存の参加者IDを使用
				QuizID:         1, // 既存のクイズIDを使用
				SelectedOption: "A",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Submit answer with invalid selected option",
			requestBody: models.AnswerRequest{
				ParticipantID:  1,
				QuizID:         1,
				SelectedOption: "E",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Submit answer with missing participant ID",
			requestBody: map[string]interface{}{
				"quiz_id":         1,
				"selected_option": "A",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Submit answer with malformed JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req, _ := http.NewRequest("POST", "/answers", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			SubmitAnswer(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestUpdateAnswer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// テスト環境用のデータベース設定
	setupTestEnv()

	// データベース接続を初期化
	_, err := database.Initialize()
	if err != nil && os.Getenv("TEST_ENV") != testEnvValue {
		t.Skipf("Database connection failed (not in test environment): %v", err)
	} else if err != nil {
		t.Fatalf("Database connection failed in test environment: %v", err)
	}

	// テスト用データを作成
	db := database.GetDB()

	// 参加者を作成
	_, err = db.Exec(`
		INSERT INTO participants (id, nickname, created_at)
		VALUES (1, 'TestUser1', CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			nickname = 'TestUser1',
			created_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		t.Fatalf("Failed to create test participant: %v", err)
	}

	// クイズを作成
	_, err = db.Exec(`
		INSERT INTO quizzes (id, question_text, option_a, option_b, option_c, option_d, correct_answer, created_at, updated_at)
		VALUES (1, 'Test Question?', 'Option A', 'Option B', 'Option C', 'Option D', 'A', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			question_text = 'Test Question?',
			option_a = 'Option A',
			option_b = 'Option B',
			option_c = 'Option C',
			option_d = 'Option D',
			correct_answer = 'A',
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		t.Fatalf("Failed to create test quiz: %v", err)
	}

	// テスト用の回答を作成（既存の回答を削除してから作成）
	_, err = db.Exec(`DELETE FROM answers WHERE participant_id = 1 AND quiz_id = 1`)
	if err != nil {
		t.Fatalf("Failed to clear test answers: %v", err)
	}
	var answerID int64
	err = db.QueryRow(`
		INSERT INTO answers (participant_id, quiz_id, selected_option, is_correct, answered_at)
		VALUES (1, 1, 'A', true, CURRENT_TIMESTAMP)
		RETURNING id
	`).Scan(&answerID)
	if err != nil {
		t.Fatalf("Failed to create test answer: %v", err)
	}

	tests := []struct {
		name           string
		answerID       string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:     "Update answer with valid data",
			answerID: fmt.Sprintf("%d", answerID),
			requestBody: models.AnswerUpdateRequest{
				SelectedOption: "B",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Update answer with invalid ID",
			answerID: "invalid",
			requestBody: models.AnswerUpdateRequest{
				SelectedOption: "A",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Update answer with invalid selected option",
			answerID: "1",
			requestBody: models.AnswerUpdateRequest{
				SelectedOption: "E",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Update answer with non-existent ID",
			answerID: "999999",
			requestBody: models.AnswerUpdateRequest{
				SelectedOption: "A",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{
				{Key: "id", Value: tt.answerID},
			}

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req, _ := http.NewRequest("PUT", "/answers/"+tt.answerID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			UpdateAnswer(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
