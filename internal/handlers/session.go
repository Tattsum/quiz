package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
)

// GetSessionStatus returns current session status
func GetSessionStatus(c *gin.Context) {
	db := database.GetDB()

	// Get current session
	var session models.QuizSession
	sessionQuery := `SELECT id, current_quiz_id, is_accepting_answers, created_at, updated_at 
					 FROM quiz_sessions 
					 ORDER BY id DESC 
					 LIMIT 1`

	err := db.QueryRow(sessionQuery).Scan(
		&session.ID,
		&session.CurrentQuizID,
		&session.IsAcceptingAnswers,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, models.APIResponse{
				Success: true,
				Data: models.SessionStatusResponse{
					SessionID:          0,
					CurrentQuiz:        nil,
					IsAcceptingAnswers: false,
					TotalParticipants:  0,
					AnswersCount:       0,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query session",
			},
		})
		return
	}

	var response models.SessionStatusResponse
	response.SessionID = session.ID
	response.IsAcceptingAnswers = session.IsAcceptingAnswers

	// Get current quiz if available
	if session.CurrentQuizID != nil {
		var quiz models.Quiz
		quizQuery := `SELECT id, question_text, option_a, option_b, option_c, option_d, 
					  image_url, video_url
					  FROM quizzes WHERE id = $1`

		err = db.QueryRow(quizQuery, *session.CurrentQuizID).Scan(
			&quiz.ID,
			&quiz.QuestionText,
			&quiz.OptionA,
			&quiz.OptionB,
			&quiz.OptionC,
			&quiz.OptionD,
			&quiz.ImageURL,
			&quiz.VideoURL,
		)

		if err == nil {
			currentQuiz := convertQuizToPublic(quiz)
			response.CurrentQuiz = &currentQuiz
		}

		// Get answers count for current quiz
		var answersCount int
		err = db.QueryRow("SELECT COUNT(*) FROM answers WHERE quiz_id = $1", *session.CurrentQuizID).Scan(&answersCount)
		if err == nil {
			response.AnswersCount = answersCount
		}
	}

	// Get total participants
	var totalParticipants int
	err = db.QueryRow("SELECT COUNT(*) FROM participants").Scan(&totalParticipants)
	if err == nil {
		response.TotalParticipants = totalParticipants
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// StartSession starts a new quiz session
func StartSession(c *gin.Context) {
	var req models.SessionStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request data",
				Details: parseValidationErrors(err),
			},
		})
		return
	}

	db := database.GetDB()

	// Check if quiz exists
	var quiz models.Quiz
	quizQuery := `SELECT id, question_text, option_a, option_b, option_c, option_d, 
				  image_url, video_url
				  FROM quizzes WHERE id = $1`

	err := db.QueryRow(quizQuery, req.QuizID).Scan(
		&quiz.ID,
		&quiz.QuestionText,
		&quiz.OptionA,
		&quiz.OptionB,
		&quiz.OptionC,
		&quiz.OptionD,
		&quiz.ImageURL,
		&quiz.VideoURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "QUIZ_NOT_FOUND",
					Message: "Quiz not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query quiz",
			},
		})
		return
	}

	// Create new session
	sessionQuery := `INSERT INTO quiz_sessions (current_quiz_id, is_accepting_answers, created_at, updated_at)
					 VALUES ($1, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
					 RETURNING id, created_at, updated_at`

	var sessionID int64
	err = db.QueryRow(sessionQuery, req.QuizID).Scan(&sessionID, &quiz.CreatedAt, &quiz.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to create session",
			},
		})
		return
	}

	currentQuiz := convertQuizToPublic(quiz)

	// Broadcast session start and first question
	BroadcastQuestionSwitch(quiz.ID, 1, 1)
	BroadcastSessionUpdate(map[string]interface{}{
		"session_id":          sessionID,
		"quiz":                currentQuiz,
		"is_accepting_answers": true,
		"status":              "started",
	})

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "クイズセッションが開始されました",
		Data: map[string]interface{}{
			"session_id":          sessionID,
			"quiz":                currentQuiz,
			"is_accepting_answers": true,
		},
	})
}

// NextQuestion moves to the next question
func NextQuestion(c *gin.Context) {
	var req models.SessionNextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request data",
				Details: parseValidationErrors(err),
			},
		})
		return
	}

	db := database.GetDB()

	// Check if quiz exists
	var quiz models.Quiz
	quizQuery := `SELECT id, question_text, option_a, option_b, option_c, option_d, 
				  image_url, video_url
				  FROM quizzes WHERE id = $1`

	err := db.QueryRow(quizQuery, req.QuizID).Scan(
		&quiz.ID,
		&quiz.QuestionText,
		&quiz.OptionA,
		&quiz.OptionB,
		&quiz.OptionC,
		&quiz.OptionD,
		&quiz.ImageURL,
		&quiz.VideoURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "QUIZ_NOT_FOUND",
					Message: "Quiz not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query quiz",
			},
		})
		return
	}

	// Get current session and update it
	sessionQuery := `UPDATE quiz_sessions 
					 SET current_quiz_id = $1, is_accepting_answers = true, updated_at = CURRENT_TIMESTAMP
					 WHERE id = (SELECT id FROM quiz_sessions ORDER BY id DESC LIMIT 1)
					 RETURNING id`

	var sessionID int64
	err = db.QueryRow(sessionQuery, req.QuizID).Scan(&sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to update session",
			},
		})
		return
	}

	currentQuiz := convertQuizToPublic(quiz)

	// Broadcast question switch (assuming question numbers for now)
	BroadcastQuestionSwitch(quiz.ID, 1, 1)
	BroadcastSessionUpdate(map[string]interface{}{
		"session_id":          sessionID,
		"quiz":                currentQuiz,
		"is_accepting_answers": true,
		"status":              "question_changed",
	})

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "次の問題に進みました",
		Data: map[string]interface{}{
			"session_id":          sessionID,
			"quiz":                currentQuiz,
			"is_accepting_answers": true,
		},
	})
}

// ToggleAnswers toggles answer acceptance for current session
func ToggleAnswers(c *gin.Context) {
	var req models.ToggleAnswersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request data",
				Details: parseValidationErrors(err),
			},
		})
		return
	}

	db := database.GetDB()

	// Update current session
	sessionQuery := `UPDATE quiz_sessions 
					 SET is_accepting_answers = $1, updated_at = CURRENT_TIMESTAMP
					 WHERE id = (SELECT id FROM quiz_sessions ORDER BY id DESC LIMIT 1)`

	_, err := db.Exec(sessionQuery, req.IsAcceptingAnswers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to update session",
			},
		})
		return
	}

	message := "回答受付を開始しました"
	if !req.IsAcceptingAnswers {
		message = "回答受付を停止しました"
		// Broadcast voting end when answers are stopped
		var currentQuizID *int64
		db.QueryRow("SELECT current_quiz_id FROM quiz_sessions ORDER BY id DESC LIMIT 1").Scan(&currentQuizID)
		if currentQuizID != nil {
			BroadcastVotingEnd(*currentQuizID, *currentQuizID)
		}
	}

	// Broadcast session update
	BroadcastSessionUpdate(map[string]interface{}{
		"is_accepting_answers": req.IsAcceptingAnswers,
		"status":              "answer_acceptance_toggled",
	})

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"is_accepting_answers": req.IsAcceptingAnswers,
		},
	})
}

// EndSession ends the current quiz session
func EndSession(c *gin.Context) {
	db := database.GetDB()

	// Update current session to stop accepting answers
	sessionQuery := `UPDATE quiz_sessions 
					 SET is_accepting_answers = false, current_quiz_id = NULL, updated_at = CURRENT_TIMESTAMP
					 WHERE id = (SELECT id FROM quiz_sessions ORDER BY id DESC LIMIT 1)`

	_, err := db.Exec(sessionQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to end session",
			},
		})
		return
	}

	// Broadcast session end
	BroadcastSessionUpdate(map[string]interface{}{
		"is_accepting_answers": false,
		"current_quiz":         nil,
		"status":               "ended",
	})

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "クイズセッションが終了されました",
	})
}