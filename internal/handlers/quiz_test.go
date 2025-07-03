package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
	"github.com/gin-gonic/gin"
)

func TestGetQuizzes(t *testing.T) {
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
		queryParams    string
		expectedStatus int
		expectedData   bool
	}{
		{
			name:           "Get quizzes with default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:           "Get quizzes with custom pagination",
			queryParams:    "?page=2&limit=5",
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest("GET", "/quizzes"+tt.queryParams, nil)
			c.Request = req

			GetQuizzes(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response models.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to unmarshal response: %v", err)
			}

			if tt.expectedData && response.Data == nil {
				t.Error("Expected data in response, but got nil")
			}
		})
	}
}

func TestGetQuiz(t *testing.T) {
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
		quizID         string
		expectedStatus int
	}{
		{
			name:           "Get quiz with valid ID",
			quizID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get quiz with invalid ID",
			quizID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get quiz with non-existent ID",
			quizID:         "999999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{
				{Key: "id", Value: tt.quizID},
			}

			GetQuiz(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestCreateQuiz(t *testing.T) {
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
			name: "Create quiz with valid data",
			requestBody: models.QuizRequest{
				QuestionText:  "What is 2+2?",
				OptionA:       "3",
				OptionB:       "4",
				OptionC:       "5",
				OptionD:       "6",
				CorrectAnswer: "B",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Create quiz with invalid data",
			requestBody: map[string]interface{}{
				"question_text": "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Create quiz with malformed JSON",
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

			req, _ := http.NewRequest("POST", "/quizzes", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			CreateQuiz(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestUpdateQuiz(t *testing.T) {
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
		quizID         string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:   "Update quiz with valid data",
			quizID: "1",
			requestBody: models.QuizRequest{
				QuestionText:  "What is 3+3?",
				OptionA:       "5",
				OptionB:       "6",
				OptionC:       "7",
				OptionD:       "8",
				CorrectAnswer: "B",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Update quiz with invalid ID",
			quizID: "invalid",
			requestBody: models.QuizRequest{
				QuestionText: "Test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Update quiz with non-existent ID",
			quizID: "999999",
			requestBody: models.QuizRequest{
				QuestionText:  "Test Question",
				OptionA:       "A",
				OptionB:       "B",
				OptionC:       "C",
				OptionD:       "D",
				CorrectAnswer: "A",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{
				{Key: "id", Value: tt.quizID},
			}

			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req, _ := http.NewRequest("PUT", "/quizzes/"+tt.quizID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			UpdateQuiz(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDeleteQuiz(t *testing.T) {
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
		quizID         string
		expectedStatus int
	}{
		{
			name:           "Delete quiz with valid ID",
			quizID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Delete quiz with invalid ID",
			quizID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Delete quiz with non-existent ID",
			quizID:         "999999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{
				{Key: "id", Value: tt.quizID},
			}

			DeleteQuiz(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
