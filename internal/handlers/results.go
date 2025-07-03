package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/Tattsum/quiz/internal/database"
	"github.com/Tattsum/quiz/internal/models"
	"github.com/gin-gonic/gin"
)

// GetCurrentResults returns results for the current active quiz
func GetCurrentResults(c *gin.Context) {
	db := database.GetDB()

	// Get current session
	var currentQuizID *int64
	var isAcceptingAnswers bool
	sessionQuery := `SELECT current_quiz_id, is_accepting_answers 
					 FROM quiz_sessions 
					 ORDER BY id DESC 
					 LIMIT 1`

	err := db.QueryRow(sessionQuery).Scan(&currentQuizID, &isAcceptingAnswers)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NO_ACTIVE_SESSION",
				Message: "No active session found",
			},
		})
		return
	}

	if currentQuizID == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NO_CURRENT_QUIZ",
				Message: "No current quiz in session",
			},
		})
		return
	}

	results, err := getQuizResultsData(db, *currentQuizID, &isAcceptingAnswers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to get quiz results",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    results,
	})
}

// GetQuizResults returns results for a specific quiz
func GetQuizResults(c *gin.Context) {
	idStr := c.Param("id")
	quizID, err := strconv.ParseInt(idStr, 10, 64)
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

	results, err := getQuizResultsData(db, quizID, nil)
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
				Message: "Failed to get quiz results",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    results,
	})
}

// getQuizResultsData retrieves and calculates quiz results
func getQuizResultsData(db *sql.DB, quizID int64, isAcceptingAnswers *bool) (*models.QuizResultsResponse, error) {
	// Get quiz info
	var questionText, correctAnswer string
	quizQuery := `SELECT question_text, correct_answer FROM quizzes WHERE id = $1`
	err := db.QueryRow(quizQuery, quizID).Scan(&questionText, &correctAnswer)
	if err != nil {
		return nil, err
	}

	// Get answer counts by option
	resultsQuery := `SELECT selected_option, COUNT(*) 
					 FROM answers 
					 WHERE quiz_id = $1 
					 GROUP BY selected_option`

	rows, err := db.Query(resultsQuery, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	optionCounts := make(map[string]int)
	totalAnswers := 0

	for rows.Next() {
		var option string
		var count int
		if err := rows.Scan(&option, &count); err != nil {
			return nil, err
		}
		optionCounts[option] = count
		totalAnswers += count
	}

	// Calculate results for each option
	results := make(map[string]models.OptionResult)
	for _, option := range []string{"A", "B", "C", "D"} {
		count := optionCounts[option]
		percentage := calculatePercentage(count, totalAnswers)
		results[option] = models.OptionResult{
			Count:      count,
			Percentage: percentage,
		}
	}

	// Get correct answer count
	correctCount := optionCounts[correctAnswer]
	correctPercentage := calculatePercentage(correctCount, totalAnswers)

	response := &models.QuizResultsResponse{
		QuizID:             quizID,
		QuestionText:       questionText,
		TotalAnswers:       totalAnswers,
		Results:            results,
		CorrectAnswer:      correctAnswer,
		CorrectCount:       correctCount,
		CorrectPercentage:  correctPercentage,
		IsAcceptingAnswers: isAcceptingAnswers,
		UpdatedAt:          time.Now(),
	}

	return response, nil
}

// GetOverallRanking returns overall participant ranking
func GetOverallRanking(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	db := database.GetDB()

	// Get total participants count
	var totalParticipants int
	err = db.QueryRow("SELECT COUNT(*) FROM participants").Scan(&totalParticipants)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to count participants",
			},
		})
		return
	}

	// Get ranking data
	rankingQuery := `SELECT p.id, p.nickname,
					 COUNT(a.id) as total_answers,
					 COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as correct_answers,
					 CASE 
						WHEN COUNT(a.id) > 0 THEN 
							CAST(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END) AS FLOAT) / COUNT(a.id)
						ELSE 0 
					 END as accuracy_rate,
					 COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as total_score
					 FROM participants p
					 LEFT JOIN answers a ON p.id = a.participant_id
					 GROUP BY p.id, p.nickname
					 ORDER BY total_score DESC, accuracy_rate DESC, total_answers DESC
					 LIMIT $1 OFFSET $2`

	rows, err := db.Query(rankingQuery, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query ranking",
			},
		})
		return
	}
	defer rows.Close()

	var ranking []models.RankingEntry
	rank := offset + 1

	for rows.Next() {
		var entry models.RankingEntry
		err := rows.Scan(
			&entry.ParticipantID,
			&entry.Nickname,
			&entry.TotalAnswers,
			&entry.CorrectAnswers,
			&entry.AccuracyRate,
			&entry.TotalScore,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "SCAN_ERROR",
					Message: "Failed to scan ranking data",
				},
			})
			return
		}
		entry.Rank = rank
		ranking = append(ranking, entry)
		rank++
	}

	response := models.OverallRankingResponse{
		Ranking:           ranking,
		TotalParticipants: totalParticipants,
		UpdatedAt:         time.Now(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetQuizRanking returns ranking for a specific quiz (correct answers)
func GetQuizRanking(c *gin.Context) {
	idStr := c.Param("id")
	quizID, err := strconv.ParseInt(idStr, 10, 64)
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

	// Get quiz info
	var questionText string
	quizQuery := `SELECT question_text FROM quizzes WHERE id = $1`
	err = db.QueryRow(quizQuery, quizID).Scan(&questionText)
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

	// Get correct participants
	correctParticipantsQuery := `SELECT p.id, p.nickname, a.selected_option, a.answered_at
								 FROM answers a
								 JOIN participants p ON a.participant_id = p.id
								 WHERE a.quiz_id = $1 AND a.is_correct = true
								 ORDER BY a.answered_at ASC`

	rows, err := db.Query(correctParticipantsQuery, quizID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to query correct participants",
			},
		})
		return
	}
	defer rows.Close()

	var correctParticipants []models.CorrectParticipant
	for rows.Next() {
		var participant models.CorrectParticipant
		err := rows.Scan(
			&participant.ParticipantID,
			&participant.Nickname,
			&participant.SelectedOption,
			&participant.AnsweredAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "SCAN_ERROR",
					Message: "Failed to scan participant data",
				},
			})
			return
		}
		correctParticipants = append(correctParticipants, participant)
	}

	// Get total answer counts
	var totalCorrect, totalAnswers int
	countQuery := `SELECT 
					COUNT(CASE WHEN is_correct THEN 1 END) as correct_count,
					COUNT(*) as total_count
					FROM answers WHERE quiz_id = $1`

	err = db.QueryRow(countQuery, quizID).Scan(&totalCorrect, &totalAnswers)
	if err != nil {
		totalCorrect = len(correctParticipants)
		totalAnswers = totalCorrect
	}

	correctPercentage := calculatePercentage(totalCorrect, totalAnswers)

	response := models.QuizRankingResponse{
		QuizID:              quizID,
		QuestionText:        questionText,
		CorrectParticipants: correctParticipants,
		TotalCorrect:        totalCorrect,
		TotalAnswers:        totalAnswers,
		CorrectPercentage:   correctPercentage,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetParticipantRanking returns ranking information for a specific participant
func GetParticipantRanking(c *gin.Context) {
	idStr := c.Param("id")
	participantID, err := strconv.ParseInt(idStr, 10, 64)
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

	// Check if participant exists and get basic info
	var nickname string
	err = db.QueryRow("SELECT nickname FROM participants WHERE id = $1", participantID).Scan(&nickname)
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
				Message: "Failed to get participant",
			},
		})
		return
	}

	// Get participant stats
	var totalAnswers, correctAnswers int
	statsQuery := `SELECT COUNT(*), COALESCE(SUM(CASE WHEN is_correct THEN 1 ELSE 0 END), 0)
				   FROM answers WHERE participant_id = $1`

	err = db.QueryRow(statsQuery, participantID).Scan(&totalAnswers, &correctAnswers)
	if err != nil {
		totalAnswers = 0
		correctAnswers = 0
	}

	var accuracyRate float64
	if totalAnswers > 0 {
		accuracyRate = float64(correctAnswers) / float64(totalAnswers)
	}

	// Get participant's rank
	rankQuery := `SELECT COUNT(*) + 1 as rank
				  FROM (
					  SELECT p.id,
						     COALESCE(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END), 0) as score,
						     CASE 
							    WHEN COUNT(a.id) > 0 THEN 
								    CAST(SUM(CASE WHEN a.is_correct THEN 1 ELSE 0 END) AS FLOAT) / COUNT(a.id)
							    ELSE 0 
						     END as acc_rate,
						     COUNT(a.id) as total_ans
					  FROM participants p
					  LEFT JOIN answers a ON p.id = a.participant_id
					  GROUP BY p.id
				  ) sub
				  WHERE (sub.score > $1) 
				     OR (sub.score = $1 AND sub.acc_rate > $2)
				     OR (sub.score = $1 AND sub.acc_rate = $2 AND sub.total_ans > $3)`

	var currentRank int
	err = db.QueryRow(rankQuery, correctAnswers, accuracyRate, totalAnswers).Scan(&currentRank)
	if err != nil {
		currentRank = 1
	}

	// Get total participants
	var totalParticipants int
	err = db.QueryRow("SELECT COUNT(*) FROM participants").Scan(&totalParticipants)
	if err != nil {
		totalParticipants = 1
	}

	// Calculate percentile
	percentile := float64(totalParticipants-currentRank+1) / float64(totalParticipants) * 100

	response := models.ParticipantRankingResponse{
		ParticipantID:     participantID,
		Nickname:          nickname,
		CurrentRank:       currentRank,
		TotalParticipants: totalParticipants,
		TotalAnswers:      totalAnswers,
		CorrectAnswers:    correctAnswers,
		AccuracyRate:      accuracyRate,
		TotalScore:        correctAnswers,
		Percentile:        percentile,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}
