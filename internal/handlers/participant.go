package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
	"github.com/gin-gonic/gin"
)

// RegisterParticipant registers a new participant
func RegisterParticipant(c *gin.Context) {
	var req models.ParticipantRequest
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

	// Insert new participant
	query := `INSERT INTO participants (nickname, created_at)
			  VALUES ($1, CURRENT_TIMESTAMP)
			  RETURNING id, created_at`

	var participant models.Participant
	err := db.QueryRow(query, req.Nickname).Scan(&participant.ID, &participant.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to register participant",
			},
		})
		return
	}

	participant.Nickname = req.Nickname

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "参加者として登録されました",
		Data: map[string]interface{}{
			"participant_id": participant.ID,
			"nickname":       participant.Nickname,
			"created_at":     participant.CreatedAt,
		},
	})
}

// GetParticipant retrieves participant information
func GetParticipant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid participant ID",
			},
		})
		return
	}

	db := database.GetDB()

	// Get participant basic info
	var participant models.Participant
	participantQuery := `SELECT id, nickname, created_at FROM participants WHERE id = $1`

	err = db.QueryRow(participantQuery, id).Scan(
		&participant.ID,
		&participant.Nickname,
		&participant.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "PARTICIPANT_NOT_FOUND",
					Message: "Participant not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query participant",
			},
		})
		return
	}

	// Get participant statistics
	var totalAnswers, correctAnswers int
	statsQuery := `SELECT COUNT(*), COALESCE(SUM(CASE WHEN is_correct THEN 1 ELSE 0 END), 0)
				   FROM answers WHERE participant_id = $1`

	err = db.QueryRow(statsQuery, id).Scan(&totalAnswers, &correctAnswers)
	if err != nil {
		// If error getting stats, just return basic info
		totalAnswers = 0
		correctAnswers = 0
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":              participant.ID,
			"nickname":        participant.Nickname,
			"created_at":      participant.CreatedAt,
			"total_answers":   totalAnswers,
			"correct_answers": correctAnswers,
		},
	})
}

// GetParticipantAnswers retrieves participant's answer history
func GetParticipantAnswers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid participant ID",
			},
		})
		return
	}

	db := database.GetDB()

	// Check if participant exists
	var participantID int64
	err = db.QueryRow("SELECT id FROM participants WHERE id = $1", id).Scan(&participantID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "PARTICIPANT_NOT_FOUND",
					Message: "Participant not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to check participant",
			},
		})
		return
	}

	// Get participant answers with quiz details
	answersQuery := `SELECT a.id, a.quiz_id, q.question_text, a.selected_option, 
					 q.correct_answer, a.is_correct, a.answered_at
					 FROM answers a
					 JOIN quizzes q ON a.quiz_id = q.id
					 WHERE a.participant_id = $1
					 ORDER BY a.answered_at DESC`

	rows, err := db.Query(answersQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query answers",
			},
		})
		return
	}
	defer rows.Close()

	var answers []models.ParticipantAnswer
	for rows.Next() {
		var answer models.ParticipantAnswer
		err := rows.Scan(
			&answer.AnswerID,
			&answer.QuizID,
			&answer.QuestionText,
			&answer.SelectedOption,
			&answer.CorrectAnswer,
			&answer.IsCorrect,
			&answer.AnsweredAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "SCAN_ERROR",
					Message: "Failed to scan answer data",
				},
			})
			return
		}
		answers = append(answers, answer)
	}

	// Calculate statistics
	totalAnswers := len(answers)
	correctAnswers := 0
	for _, answer := range answers {
		if answer.IsCorrect {
			correctAnswers++
		}
	}

	var accuracyRate float64
	if totalAnswers > 0 {
		accuracyRate = float64(correctAnswers) / float64(totalAnswers)
	}

	response := models.ParticipantAnswersResponse{
		ParticipantID:  id,
		Answers:        answers,
		TotalAnswers:   totalAnswers,
		CorrectAnswers: correctAnswers,
		AccuracyRate:   accuracyRate,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// SubmitAnswer handles answer submission
func SubmitAnswer(c *gin.Context) {
	var req models.AnswerRequest
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

	// Check if session is accepting answers
	var isAcceptingAnswers bool
	var currentQuizID *int64
	sessionQuery := `SELECT is_accepting_answers, current_quiz_id 
					 FROM quiz_sessions 
					 ORDER BY id DESC 
					 LIMIT 1`

	err := db.QueryRow(sessionQuery).Scan(&isAcceptingAnswers, &currentQuizID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SESSION_ERROR",
				Message: "No active session found",
			},
		})
		return
	}

	if !isAcceptingAnswers {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ANSWERS_NOT_ACCEPTED",
				Message: "Answer submission is currently not accepted",
			},
		})
		return
	}

	if currentQuizID == nil || *currentQuizID != req.QuizID {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_QUIZ",
				Message: "This quiz is not currently active",
			},
		})
		return
	}

	// Check if participant exists
	var participantID int64
	err = db.QueryRow("SELECT id FROM participants WHERE id = $1", req.ParticipantID).Scan(&participantID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "PARTICIPANT_NOT_FOUND",
					Message: "Participant not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to check participant",
			},
		})
		return
	}

	// Get quiz correct answer
	var correctAnswer string
	err = db.QueryRow("SELECT correct_answer FROM quizzes WHERE id = $1", req.QuizID).Scan(&correctAnswer)
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
				Message: "Failed to get quiz",
			},
		})
		return
	}

	isCorrect := req.SelectedOption == correctAnswer

	// Check if answer already exists (for update)
	var existingAnswerID int64
	checkQuery := `SELECT id FROM answers WHERE participant_id = $1 AND quiz_id = $2`
	err = db.QueryRow(checkQuery, req.ParticipantID, req.QuizID).Scan(&existingAnswerID)

	if err == nil {
		// Update existing answer
		updateQuery := `UPDATE answers 
						SET selected_option = $1, is_correct = $2, answered_at = CURRENT_TIMESTAMP
						WHERE id = $3
						RETURNING id, answered_at`

		var answer models.Answer
		err = db.QueryRow(updateQuery, req.SelectedOption, isCorrect, existingAnswerID).Scan(
			&answer.ID, &answer.AnsweredAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "Failed to update answer",
				},
			})
			return
		}

		answer.ParticipantID = req.ParticipantID
		answer.QuizID = req.QuizID
		answer.SelectedOption = req.SelectedOption
		answer.IsCorrect = isCorrect

		// Broadcast answer status update
		db := database.GetDB()
		var totalParticipants, answeredCount int
		answerCounts := make(map[string]int)

		// Get total participants
		db.QueryRow("SELECT COUNT(*) FROM participants").Scan(&totalParticipants)

		// Get answered count for this quiz
		db.QueryRow("SELECT COUNT(*) FROM answers WHERE quiz_id = $1", req.QuizID).Scan(&answeredCount)

		// Get answer distribution
		rows, err := db.Query("SELECT selected_option, COUNT(*) FROM answers WHERE quiz_id = $1 GROUP BY selected_option", req.QuizID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var option string
				var count int
				rows.Scan(&option, &count)
				answerCounts[option] = count
			}
		}

		// Broadcast the current answer status
		BroadcastAnswerStatus(req.QuizID, req.QuizID, totalParticipants, answeredCount, answerCounts)

		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "回答が変更されました",
			Data:    answer,
		})
	} else if err == sql.ErrNoRows {
		// Insert new answer
		insertQuery := `INSERT INTO answers (participant_id, quiz_id, selected_option, is_correct, answered_at)
						VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
						RETURNING id, answered_at`

		var answer models.Answer
		err = db.QueryRow(insertQuery, req.ParticipantID, req.QuizID, req.SelectedOption, isCorrect).Scan(
			&answer.ID, &answer.AnsweredAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "DATABASE_ERROR",
					Message: "Failed to submit answer",
				},
			})
			return
		}

		answer.ParticipantID = req.ParticipantID
		answer.QuizID = req.QuizID
		answer.SelectedOption = req.SelectedOption
		answer.IsCorrect = isCorrect

		// Broadcast answer status update
		db := database.GetDB()
		var totalParticipants, answeredCount int
		answerCounts := make(map[string]int)

		// Get total participants
		db.QueryRow("SELECT COUNT(*) FROM participants").Scan(&totalParticipants)

		// Get answered count for this quiz
		db.QueryRow("SELECT COUNT(*) FROM answers WHERE quiz_id = $1", req.QuizID).Scan(&answeredCount)

		// Get answer distribution
		rows, err := db.Query("SELECT selected_option, COUNT(*) FROM answers WHERE quiz_id = $1 GROUP BY selected_option", req.QuizID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var option string
				var count int
				rows.Scan(&option, &count)
				answerCounts[option] = count
			}
		}

		// Broadcast the current answer status
		BroadcastAnswerStatus(req.QuizID, req.QuizID, totalParticipants, answeredCount, answerCounts)

		c.JSON(http.StatusCreated, models.APIResponse{
			Success: true,
			Message: "回答が送信されました",
			Data:    answer,
		})
	} else {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to check existing answer",
			},
		})
		return
	}
}

// UpdateAnswer handles answer updates
func UpdateAnswer(c *gin.Context) {
	idStr := c.Param("id")
	answerID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid answer ID",
			},
		})
		return
	}

	var req models.AnswerUpdateRequest
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

	// Check if session is accepting answers
	var isAcceptingAnswers bool
	sessionQuery := `SELECT is_accepting_answers FROM quiz_sessions ORDER BY id DESC LIMIT 1`
	err = db.QueryRow(sessionQuery).Scan(&isAcceptingAnswers)
	if err != nil || !isAcceptingAnswers {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "ANSWERS_NOT_ACCEPTED",
				Message: "Answer updates are not currently accepted",
			},
		})
		return
	}

	// Get existing answer and quiz info
	var quizID int64
	var correctAnswer string
	existingQuery := `SELECT a.quiz_id, q.correct_answer 
					  FROM answers a 
					  JOIN quizzes q ON a.quiz_id = q.id 
					  WHERE a.id = $1`

	err = db.QueryRow(existingQuery, answerID).Scan(&quizID, &correctAnswer)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "ANSWER_NOT_FOUND",
					Message: "Answer not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to get answer",
			},
		})
		return
	}

	isCorrect := req.SelectedOption == correctAnswer

	// Update answer
	updateQuery := `UPDATE answers 
					SET selected_option = $1, is_correct = $2, answered_at = CURRENT_TIMESTAMP
					WHERE id = $3
					RETURNING participant_id, quiz_id, answered_at`

	var answer models.Answer
	err = db.QueryRow(updateQuery, req.SelectedOption, isCorrect, answerID).Scan(
		&answer.ParticipantID, &answer.QuizID, &answer.AnsweredAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to update answer",
			},
		})
		return
	}

	answer.ID = answerID
	answer.SelectedOption = req.SelectedOption
	answer.IsCorrect = isCorrect

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "回答が変更されました",
		Data:    answer,
	})
}
