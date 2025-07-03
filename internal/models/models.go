// Package models contains data structures and domain models for the quiz application.
package models

import (
	"time"
)

// Administrator represents the administrators table
type Administrator struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Email        string    `json:"email" db:"email"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Participant represents the participants table
type Participant struct {
	ID        int64     `json:"id" db:"id"`
	Nickname  string    `json:"nickname" db:"nickname"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Quiz represents the quizzes table
type Quiz struct {
	ID            int64     `json:"id" db:"id"`
	QuestionText  string    `json:"question_text" db:"question_text"`
	OptionA       string    `json:"option_a" db:"option_a"`
	OptionB       string    `json:"option_b" db:"option_b"`
	OptionC       string    `json:"option_c" db:"option_c"`
	OptionD       string    `json:"option_d" db:"option_d"`
	CorrectAnswer string    `json:"correct_answer,omitempty" db:"correct_answer"`
	ImageURL      *string   `json:"image_url" db:"image_url"`
	VideoURL      *string   `json:"video_url" db:"video_url"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// QuizPublic represents a quiz without the correct answer (for public display)
type QuizPublic struct {
	ID           int64   `json:"id"`
	QuestionText string  `json:"question_text"`
	OptionA      string  `json:"option_a"`
	OptionB      string  `json:"option_b"`
	OptionC      string  `json:"option_c"`
	OptionD      string  `json:"option_d"`
	ImageURL     *string `json:"image_url"`
	VideoURL     *string `json:"video_url"`
}

// Answer represents the answers table
type Answer struct {
	ID             int64     `json:"id" db:"id"`
	ParticipantID  int64     `json:"participant_id" db:"participant_id"`
	QuizID         int64     `json:"quiz_id" db:"quiz_id"`
	SelectedOption string    `json:"selected_option" db:"selected_option"`
	IsCorrect      bool      `json:"is_correct" db:"is_correct"`
	AnsweredAt     time.Time `json:"answered_at" db:"answered_at"`
}

// QuizSession represents the quiz_sessions table
type QuizSession struct {
	ID                 int64     `json:"id" db:"id"`
	CurrentQuizID      *int64    `json:"current_quiz_id" db:"current_quiz_id"`
	IsAcceptingAnswers bool      `json:"is_accepting_answers" db:"is_accepting_answers"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// Request/Response DTOs

// LoginRequest represents admin login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents admin login response
type LoginResponse struct {
	Token     string        `json:"token"`
	ExpiresAt time.Time     `json:"expires_at"`
	Admin     Administrator `json:"admin"`
}

// QuizRequest represents quiz creation/update request
type QuizRequest struct {
	QuestionText  string  `json:"question_text" binding:"required"`
	OptionA       string  `json:"option_a" binding:"required"`
	OptionB       string  `json:"option_b" binding:"required"`
	OptionC       string  `json:"option_c" binding:"required"`
	OptionD       string  `json:"option_d" binding:"required"`
	CorrectAnswer string  `json:"correct_answer" binding:"required,oneof=A B C D"`
	ImageURL      *string `json:"image_url"`
	VideoURL      *string `json:"video_url"`
}

// ParticipantRequest represents participant registration request
type ParticipantRequest struct {
	Nickname string `json:"nickname" binding:"required,max=50"`
}

// AnswerRequest represents answer submission request
type AnswerRequest struct {
	ParticipantID  int64  `json:"participant_id" binding:"required"`
	QuizID         int64  `json:"quiz_id" binding:"required"`
	SelectedOption string `json:"selected_option" binding:"required,oneof=A B C D"`
}

// AnswerUpdateRequest represents answer update request
type AnswerUpdateRequest struct {
	SelectedOption string `json:"selected_option" binding:"required,oneof=A B C D"`
}

// SessionStartRequest represents session start request
type SessionStartRequest struct {
	QuizID int64 `json:"quiz_id" binding:"required"`
}

// SessionNextRequest represents next question request
type SessionNextRequest struct {
	QuizID int64 `json:"quiz_id" binding:"required"`
}

// ToggleAnswersRequest represents toggle answers request
type ToggleAnswersRequest struct {
	IsAcceptingAnswers bool `json:"is_accepting_answers"`
}

// SessionStatusResponse represents session status response
type SessionStatusResponse struct {
	SessionID          int64       `json:"session_id"`
	CurrentQuiz        *QuizPublic `json:"current_quiz"`
	IsAcceptingAnswers bool        `json:"is_accepting_answers"`
	TotalParticipants  int         `json:"total_participants"`
	AnswersCount       int         `json:"answers_count"`
}

// QuizResultsResponse represents quiz results response
type QuizResultsResponse struct {
	QuizID             int64                   `json:"quiz_id"`
	QuestionText       string                  `json:"question_text"`
	TotalAnswers       int                     `json:"total_answers"`
	Results            map[string]OptionResult `json:"results"`
	CorrectAnswer      string                  `json:"correct_answer"`
	CorrectCount       int                     `json:"correct_count"`
	CorrectPercentage  float64                 `json:"correct_percentage"`
	IsAcceptingAnswers *bool                   `json:"is_accepting_answers,omitempty"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

// OptionResult represents result for each option
type OptionResult struct {
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// RankingEntry represents a ranking entry
type RankingEntry struct {
	Rank           int     `json:"rank"`
	ParticipantID  int64   `json:"participant_id"`
	Nickname       string  `json:"nickname"`
	TotalAnswers   int     `json:"total_answers"`
	CorrectAnswers int     `json:"correct_answers"`
	AccuracyRate   float64 `json:"accuracy_rate"`
	TotalScore     int     `json:"total_score"`
}

// OverallRankingResponse represents overall ranking response
type OverallRankingResponse struct {
	Ranking           []RankingEntry `json:"ranking"`
	TotalParticipants int            `json:"total_participants"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// QuizRankingResponse represents quiz-specific ranking response
type QuizRankingResponse struct {
	QuizID              int64                `json:"quiz_id"`
	QuestionText        string               `json:"question_text"`
	CorrectParticipants []CorrectParticipant `json:"correct_participants"`
	TotalCorrect        int                  `json:"total_correct"`
	TotalAnswers        int                  `json:"total_answers"`
	CorrectPercentage   float64              `json:"correct_percentage"`
}

// CorrectParticipant represents a participant who answered correctly
type CorrectParticipant struct {
	ParticipantID  int64     `json:"participant_id"`
	Nickname       string    `json:"nickname"`
	SelectedOption string    `json:"selected_option"`
	AnsweredAt     time.Time `json:"answered_at"`
}

// ParticipantRankingResponse represents participant ranking response
type ParticipantRankingResponse struct {
	ParticipantID     int64   `json:"participant_id"`
	Nickname          string  `json:"nickname"`
	CurrentRank       int     `json:"current_rank"`
	TotalParticipants int     `json:"total_participants"`
	TotalAnswers      int     `json:"total_answers"`
	CorrectAnswers    int     `json:"correct_answers"`
	AccuracyRate      float64 `json:"accuracy_rate"`
	TotalScore        int     `json:"total_score"`
	Percentile        float64 `json:"percentile"`
}

// ParticipantAnswersResponse represents participant answers history response
type ParticipantAnswersResponse struct {
	ParticipantID  int64               `json:"participant_id"`
	Answers        []ParticipantAnswer `json:"answers"`
	TotalAnswers   int                 `json:"total_answers"`
	CorrectAnswers int                 `json:"correct_answers"`
	AccuracyRate   float64             `json:"accuracy_rate"`
}

// ParticipantAnswer represents a single answer in participant's history
type ParticipantAnswer struct {
	AnswerID       int64     `json:"answer_id"`
	QuizID         int64     `json:"quiz_id"`
	QuestionText   string    `json:"question_text"`
	SelectedOption string    `json:"selected_option"`
	CorrectAnswer  string    `json:"correct_answer"`
	IsCorrect      bool      `json:"is_correct"`
	AnsweredAt     time.Time `json:"answered_at"`
}

// APIResponse represents standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents API error details
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError represents field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// PaginatedResponse represents paginated response
type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// UploadResponse represents file upload response
type UploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// ImageUpload represents uploaded image metadata
type ImageUpload struct {
	ID           int64     `json:"id" db:"id"`
	OriginalName string    `json:"original_name" db:"original_name"`
	Filename     string    `json:"filename" db:"filename"`
	Path         string    `json:"path" db:"path"`
	URL          string    `json:"url" db:"url"`
	ContentType  string    `json:"content_type" db:"content_type"`
	Size         int64     `json:"size" db:"size"`
	Width        int       `json:"width" db:"width"`
	Height       int       `json:"height" db:"height"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ImageUploadRequest represents image upload request validation
type ImageUploadRequest struct {
	MaxFileSize   int64    `json:"max_file_size" default:"5242880"`
	AllowedTypes  []string `json:"allowed_types" default:"[\"image/jpeg\",\"image/png\",\"image/gif\"]"`
	MaxWidth      int      `json:"max_width" default:"1920"`
	ResizeQuality int      `json:"resize_quality" default:"80"`
}
