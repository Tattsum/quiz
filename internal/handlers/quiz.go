package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
)

// GetQuizzes retrieves all quizzes with pagination
func GetQuizzes(c *gin.Context) {
	page, limit, _ := getPaginationParams(c)
	offset := getOffset(page, limit)

	db := database.GetDB()

	// Get total count
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM quizzes").Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to count quizzes",
			},
		})
		return
	}

	// Get quizzes with pagination
	query := `SELECT id, question_text, option_a, option_b, option_c, option_d, 
			  correct_answer, image_url, video_url, created_at, updated_at
			  FROM quizzes 
			  ORDER BY created_at DESC 
			  LIMIT $1 OFFSET $2`

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query quizzes",
			},
		})
		return
	}
	defer rows.Close()

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
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "SCAN_ERROR",
					Message: "Failed to scan quiz data",
				},
			})
			return
		}
		quizzes = append(quizzes, quiz)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.PaginatedResponse{
			Data:  quizzes,
			Total: total,
			Page:  page,
			Limit: limit,
		},
	})
}

// GetQuiz retrieves a single quiz by ID
func GetQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid quiz ID",
			},
		})
		return
	}

	db := database.GetDB()
	
	var quiz models.Quiz
	query := `SELECT id, question_text, option_a, option_b, option_c, option_d, 
			  correct_answer, image_url, video_url, created_at, updated_at
			  FROM quizzes WHERE id = $1`

	err = db.QueryRow(query, id).Scan(
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

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    quiz,
	})
}

// CreateQuiz creates a new quiz
func CreateQuiz(c *gin.Context) {
	var req models.QuizRequest
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

	query := `INSERT INTO quizzes (question_text, option_a, option_b, option_c, option_d, 
			  correct_answer, image_url, video_url, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			  RETURNING id, created_at, updated_at`

	var quiz models.Quiz
	err := db.QueryRow(query,
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
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to create quiz",
			},
		})
		return
	}

	// Fill quiz struct with request data
	quiz.QuestionText = req.QuestionText
	quiz.OptionA = req.OptionA
	quiz.OptionB = req.OptionB
	quiz.OptionC = req.OptionC
	quiz.OptionD = req.OptionD
	quiz.CorrectAnswer = req.CorrectAnswer
	quiz.ImageURL = req.ImageURL
	quiz.VideoURL = req.VideoURL

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "問題が作成されました",
		Data:    quiz,
	})
}

// UpdateQuiz updates an existing quiz
func UpdateQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid quiz ID",
			},
		})
		return
	}

	var req models.QuizRequest
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
	var existingID int64
	err = db.QueryRow("SELECT id FROM quizzes WHERE id = $1", id).Scan(&existingID)
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
				Message: "Failed to check quiz existence",
			},
		})
		return
	}

	// Update quiz
	query := `UPDATE quizzes 
			  SET question_text = $1, option_a = $2, option_b = $3, option_c = $4, 
				  option_d = $5, correct_answer = $6, image_url = $7, video_url = $8, 
				  updated_at = CURRENT_TIMESTAMP
			  WHERE id = $9
			  RETURNING updated_at`

	var quiz models.Quiz
	err = db.QueryRow(query,
		req.QuestionText,
		req.OptionA,
		req.OptionB,
		req.OptionC,
		req.OptionD,
		req.CorrectAnswer,
		req.ImageURL,
		req.VideoURL,
		id,
	).Scan(&quiz.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to update quiz",
			},
		})
		return
	}

	// Get the updated quiz
	selectQuery := `SELECT id, question_text, option_a, option_b, option_c, option_d, 
					correct_answer, image_url, video_url, created_at, updated_at
					FROM quizzes WHERE id = $1`

	err = db.QueryRow(selectQuery, id).Scan(
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
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to retrieve updated quiz",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "問題が更新されました",
		Data:    quiz,
	})
}

// DeleteQuiz deletes a quiz by ID
func DeleteQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_ID",
				Message: "Invalid quiz ID",
			},
		})
		return
	}

	db := database.GetDB()

	// Check if quiz exists
	var existingID int64
	err = db.QueryRow("SELECT id FROM quizzes WHERE id = $1", id).Scan(&existingID)
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
				Message: "Failed to check quiz existence",
			},
		})
		return
	}

	// Delete quiz (answers will be deleted automatically due to CASCADE)
	_, err = db.Exec("DELETE FROM quizzes WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to delete quiz",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "問題が削除されました",
	})
}