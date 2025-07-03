package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
)

// QuizService provides quiz related business logic
type QuizService struct {
	db *sql.DB
}

// NewQuizService creates a new QuizService instance
func NewQuizService() *QuizService {
	return &QuizService{
		db: database.GetDB(),
	}
}

// CreateQuiz creates a new quiz in the database
func (s *QuizService) CreateQuiz(req models.QuizRequest) (*models.Quiz, error) {
	if err := s.validateQuizRequest(req); err != nil {
		return nil, err
	}

	query := `INSERT INTO quizzes (question_text, option_a, option_b, option_c, option_d, 
			  correct_answer, image_url, video_url, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			  RETURNING id, created_at, updated_at`

	var quiz models.Quiz
	err := s.db.QueryRow(query,
		req.QuestionText,
		req.OptionA,
		req.OptionB,
		req.OptionC,
		req.OptionD,
		req.CorrectAnswer,
		req.ImageURL,
		req.VideoURL,
	).Scan(&quiz.ID, &quiz.CreatedAt, &quiz.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create quiz: %w", err)
	}

	quiz.QuestionText = req.QuestionText
	quiz.OptionA = req.OptionA
	quiz.OptionB = req.OptionB
	quiz.OptionC = req.OptionC
	quiz.OptionD = req.OptionD
	quiz.CorrectAnswer = req.CorrectAnswer
	quiz.ImageURL = req.ImageURL
	quiz.VideoURL = req.VideoURL

	return &quiz, nil
}

// GetQuizByID retrieves a quiz by its ID
func (s *QuizService) GetQuizByID(id int64) (*models.Quiz, error) {
	if id <= 0 {
		return nil, errors.New("invalid quiz ID")
	}

	var quiz models.Quiz
	query := `SELECT id, question_text, option_a, option_b, option_c, option_d, 
			  correct_answer, image_url, video_url, created_at, updated_at
			  FROM quizzes WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&quiz.ID,
		&quiz.QuestionText,
		&quiz.OptionA,
		&quiz.OptionB,
		&quiz.OptionC,
		&quiz.OptionD,
		&quiz.CorrectAnswer,
		&quiz.ImageURL,
		&quiz.VideoURL,
		&quiz.CreatedAt,
		&quiz.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("quiz not found")
		}
		return nil, fmt.Errorf("failed to get quiz: %w", err)
	}

	return &quiz, nil
}

// GetPublicQuizByID retrieves a quiz by its ID without the correct answer
func (s *QuizService) GetPublicQuizByID(id int64) (*models.QuizPublic, error) {
	quiz, err := s.GetQuizByID(id)
	if err != nil {
		return nil, err
	}

	return &models.QuizPublic{
		ID:           quiz.ID,
		QuestionText: quiz.QuestionText,
		OptionA:      quiz.OptionA,
		OptionB:      quiz.OptionB,
		OptionC:      quiz.OptionC,
		OptionD:      quiz.OptionD,
		ImageURL:     quiz.ImageURL,
		VideoURL:     quiz.VideoURL,
	}, nil
}

// UpdateQuiz updates an existing quiz in the database
func (s *QuizService) UpdateQuiz(id int64, req models.QuizRequest) (*models.Quiz, error) {
	if id <= 0 {
		return nil, errors.New("invalid quiz ID")
	}

	if err := s.validateQuizRequest(req); err != nil {
		return nil, err
	}

	_, err := s.GetQuizByID(id)
	if err != nil {
		return nil, err
	}

	query := `UPDATE quizzes 
			  SET question_text = $1, option_a = $2, option_b = $3, option_c = $4, 
				  option_d = $5, correct_answer = $6, image_url = $7, video_url = $8, 
				  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $9
			  RETURNING updated_at`

	var updatedAt sql.NullTime
	err = s.db.QueryRow(query,
		req.QuestionText,
		req.OptionA,
		req.OptionB,
		req.OptionC,
		req.OptionD,
		req.CorrectAnswer,
		req.ImageURL,
		req.VideoURL,
		id,
	).Scan(&updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update quiz: %w", err)
	}

	return s.GetQuizByID(id)
}

// DeleteQuiz deletes a quiz from the database
func (s *QuizService) DeleteQuiz(id int64) error {
	if id <= 0 {
		return errors.New("invalid quiz ID")
	}

	_, err := s.GetQuizByID(id)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM quizzes WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete quiz: %w", err)
	}

	return nil
}

// GetQuizzes retrieves a paginated list of quizzes
func (s *QuizService) GetQuizzes(page, limit int) ([]models.Quiz, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM quizzes").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count quizzes: %w", err)
	}

	query := `SELECT id, question_text, option_a, option_b, option_c, option_d, 
			  correct_answer, image_url, video_url, created_at, updated_at
			  FROM quizzes 
			  ORDER BY created_at DESC 
			  LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query quizzes: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Error closing rows: %v\n", closeErr)
		}
	}()

	var quizzes []models.Quiz
	for rows.Next() {
		var quiz models.Quiz
		err := rows.Scan(
			&quiz.ID,
			&quiz.QuestionText,
			&quiz.OptionA,
			&quiz.OptionB,
			&quiz.OptionC,
			&quiz.OptionD,
			&quiz.CorrectAnswer,
			&quiz.ImageURL,
			&quiz.VideoURL,
			&quiz.CreatedAt,
			&quiz.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan quiz: %w", err)
		}
		quizzes = append(quizzes, quiz)
	}

	return quizzes, total, nil
}

func (s *QuizService) validateQuizRequest(req models.QuizRequest) error {
	if req.QuestionText == "" {
		return errors.New("question text is required")
	}
	if req.OptionA == "" || req.OptionB == "" || req.OptionC == "" || req.OptionD == "" {
		return errors.New("all options are required")
	}
	if req.CorrectAnswer != "A" && req.CorrectAnswer != "B" && req.CorrectAnswer != "C" && req.CorrectAnswer != "D" {
		return errors.New("correct answer must be A, B, C, or D")
	}
	return nil
}
