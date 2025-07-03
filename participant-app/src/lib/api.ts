import { Participant, Question, QuizSession, Answer, AnswerOption } from '@/types'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export class ApiClient {
  private baseUrl: string

  constructor(baseUrl = API_BASE_URL) {
    this.baseUrl = baseUrl
  }

  async joinQuiz(nickname: string, quizId: string): Promise<Participant> {
    const response = await fetch(`${this.baseUrl}/api/participants`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        nickname,
        quiz_id: quizId,
      }),
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Unknown error' }))
      throw new Error(error.message || 'Failed to join quiz')
    }

    return response.json()
  }

  async submitAnswer(participantId: string, questionId: string, selectedOption: number): Promise<Answer> {
    const response = await fetch(`${this.baseUrl}/api/answers`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        participant_id: participantId,
        question_id: questionId,
        selected_option: selectedOption,
      }),
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Unknown error' }))
      throw new Error(error.message || 'Failed to submit answer')
    }

    return response.json()
  }

  async getQuizSession(quizId: string): Promise<QuizSession> {
    const response = await fetch(`${this.baseUrl}/api/quiz/${quizId}/session`)

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Unknown error' }))
      throw new Error(error.message || 'Failed to get quiz session')
    }

    return response.json()
  }

  async getCurrentQuestion(quizId: string): Promise<Question> {
    const response = await fetch(`${this.baseUrl}/api/quiz/${quizId}/current-question`)

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Unknown error' }))
      throw new Error(error.message || 'Failed to get current question')
    }

    return response.json()
  }

  async getParticipantResult(participantId: string): Promise<any> {
    const response = await fetch(`${this.baseUrl}/api/participants/${participantId}/result`)

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Unknown error' }))
      throw new Error(error.message || 'Failed to get participant result')
    }

    return response.json()
  }
}

export const apiClient = new ApiClient()