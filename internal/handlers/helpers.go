package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/Tattsum/quiz/internal/models"
)

// parseValidationErrors converts validation errors to API error format
func parseValidationErrors(err error) []models.ValidationError {
	var errors []models.ValidationError
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, models.ValidationError{
				Field:   getFieldName(e.Tag(), e.Field()),
				Message: getValidationMessage(e),
			})
		}
	}
	
	return errors
}

// getFieldName returns user-friendly field name
func getFieldName(tag, field string) string {
	// Convert field names to user-friendly format
	switch strings.ToLower(field) {
	case "username":
		return "ユーザー名"
	case "password":
		return "パスワード"
	case "nickname":
		return "ニックネーム"
	case "questiontext":
		return "問題文"
	case "optiona", "optionb", "optionc", "optiond":
		return "選択肢"
	case "correctanswer":
		return "正解"
	case "participantid":
		return "参加者ID"
	case "quizid":
		return "問題ID"
	case "selectedoption":
		return "選択した答え"
	default:
		return field
	}
}

// getValidationMessage returns user-friendly validation message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "必須項目です"
	case "max":
		return "文字数が上限を超えています"
	case "min":
		return "文字数が不足しています"
	case "oneof":
		return "有効な値を選択してください"
	case "email":
		return "有効なメールアドレスを入力してください"
	default:
		return "入力値が不正です"
	}
}

// getPaginationParams extracts pagination parameters from query
func getPaginationParams(c *gin.Context) (int, int, error) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	
	return page, limit, nil
}

// getOffset calculates database offset from page and limit
func getOffset(page, limit int) int {
	return (page - 1) * limit
}

// convertQuizToPublic converts Quiz model to QuizPublic (without correct answer)
func convertQuizToPublic(quiz models.Quiz) models.QuizPublic {
	return models.QuizPublic{
		ID:           quiz.ID,
		QuestionText: quiz.QuestionText,
		OptionA:      quiz.OptionA,
		OptionB:      quiz.OptionB,
		OptionC:      quiz.OptionC,
		OptionD:      quiz.OptionD,
		ImageURL:     quiz.ImageURL,
		VideoURL:     quiz.VideoURL,
	}
}

// calculatePercentage calculates percentage with proper rounding
func calculatePercentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

// isValidOption checks if the option is A, B, C, or D
func isValidOption(option string) bool {
	return option == "A" || option == "B" || option == "C" || option == "D"
}