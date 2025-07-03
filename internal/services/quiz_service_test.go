package services

import (
	"os"
	"testing"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
)

func TestQuizService_validateQuizRequest(t *testing.T) {
	service := NewQuizService()

	tests := []struct {
		name    string
		req     models.QuizRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: models.QuizRequest{
				QuestionText:  "Test question?",
				OptionA:       "Option A",
				OptionB:       "Option B",
				OptionC:       "Option C",
				OptionD:       "Option D",
				CorrectAnswer: "A",
			},
			wantErr: false,
		},
		{
			name: "missing question text",
			req: models.QuizRequest{
				QuestionText:  "",
				OptionA:       "Option A",
				OptionB:       "Option B",
				OptionC:       "Option C",
				OptionD:       "Option D",
				CorrectAnswer: "A",
			},
			wantErr: true,
		},
		{
			name: "missing option A",
			req: models.QuizRequest{
				QuestionText:  "Test question?",
				OptionA:       "",
				OptionB:       "Option B",
				OptionC:       "Option C",
				OptionD:       "Option D",
				CorrectAnswer: "A",
			},
			wantErr: true,
		},
		{
			name: "invalid correct answer",
			req: models.QuizRequest{
				QuestionText:  "Test question?",
				OptionA:       "Option A",
				OptionB:       "Option B",
				OptionC:       "Option C",
				OptionD:       "Option D",
				CorrectAnswer: "E",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateQuizRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateQuizRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQuizService_GetQuizByID(t *testing.T) {
	// テスト環境設定
	if os.Getenv("TEST_ENV") != "true" {
		// Set default test database configuration
		_ = os.Setenv("DB_HOST", "localhost")
		_ = os.Setenv("DB_PORT", "5433")
		_ = os.Setenv("DB_USER", "quiz_user")
		_ = os.Setenv("DB_PASSWORD", "quiz_password")
		_ = os.Setenv("DB_NAME", "quiz_db_test")
		_ = os.Setenv("DB_SSLMODE", "disable")
	}

	// データベース接続を初期化
	_, err := database.Initialize()
	if err != nil {
		t.Skipf("Database connection failed: %v", err)
	}

	service := NewQuizService()

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "invalid ID - zero",
			id:      0,
			wantErr: true,
		},
		{
			name:    "invalid ID - negative",
			id:      -1,
			wantErr: true,
		},
		{
			name:    "valid ID format",
			id:      1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetQuizByID(tt.id)
			if tt.name == "valid ID format" {
				if err != nil && err.Error() != "quiz not found" {
					t.Errorf("GetQuizByID() unexpected error = %v", err)
				}
			} else if (err != nil) != tt.wantErr {
				t.Errorf("GetQuizByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
