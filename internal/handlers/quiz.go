package handlers

import (
	"net/http"
	"strconv"

	"github.com/Tattsum/quiz/internal/models"
	"github.com/Tattsum/quiz/internal/services"
	"github.com/gin-gonic/gin"
)

const quizNotFoundError = "quiz not found"

// GetQuizzes retrieves all quizzes with pagination
func GetQuizzes(c *gin.Context) {
	page, limit, _ := getPaginationParams(c)

	quizService := services.NewQuizService()
	quizzes, total, err := quizService.GetQuizzes(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to retrieve quizzes",
			},
		})
		return
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

	quizService := services.NewQuizService()
	quiz, err := quizService.GetQuizByID(id)
	if err != nil {
		if err.Error() == quizNotFoundError {
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

	quizService := services.NewQuizService()
	quiz, err := quizService.CreateQuiz(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

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

	quizService := services.NewQuizService()
	quiz, err := quizService.UpdateQuiz(id, req)
	if err != nil {
		if err.Error() == quizNotFoundError {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "QUIZ_NOT_FOUND",
					Message: "Quiz not found",
				},
			})
			return
		}
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
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

	quizService := services.NewQuizService()
	err = quizService.DeleteQuiz(id)
	if err != nil {
		if err.Error() == quizNotFoundError {
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
