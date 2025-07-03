export interface Participant {
  id: string
  nickname: string
  quizId: string
  createdAt: string
}

export interface Question {
  id: string
  text: string
  options: string[]
  correctAnswer: number
  order: number
}

export interface Answer {
  participantId: string
  questionId: string
  selectedOption: number
  submittedAt: string
}

export interface QuizSession {
  id: string
  title: string
  currentQuestionNumber: number
  totalQuestions: number
  status: 'waiting' | 'question' | 'voting_ended' | 'finished'
}

export interface AnswerStats {
  totalParticipants: number
  answeredCount: number
  answerCounts: {
    A: number
    B: number
    C: number
    D: number
  }
}

export interface WebSocketMessage {
  type: 'question_switch' | 'voting_end' | 'answer_status' | 'session_update' | 'result_update'
  data: any
}

export type AnswerOption = 'A' | 'B' | 'C' | 'D'

export interface QuestionSwitchData {
  questionNumber: number
  totalQuestions: number
  question: Question
}

export interface VotingEndData {
  questionId: string
}

export interface SessionUpdateData {
  session: QuizSession
}