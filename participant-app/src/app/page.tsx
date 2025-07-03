'use client'

import { useState, useEffect, useCallback } from 'react'
import { useSearchParams } from 'next/navigation'
import NicknameInput from '@/components/NicknameInput'
import AnswerScreen from '@/components/AnswerScreen'
import WaitingScreen from '@/components/WaitingScreen'
import ResultScreen from '@/components/ResultScreen'
import { apiClient } from '@/lib/api'
import { wsClient } from '@/lib/websocket'
import { 
  Participant, 
  Question, 
  AnswerOption, 
  QuizSession, 
  WebSocketMessage,
  QuestionSwitchData,
  VotingEndData,
  SessionUpdateData
} from '@/types'

type AppState = 'nickname' | 'waiting' | 'question' | 'voting_ended' | 'result'

export default function HomePage() {
  const searchParams = useSearchParams()
  const quizId = searchParams.get('quiz') || '1' // デフォルトのクイズID

  const [state, setState] = useState<AppState>('nickname')
  const [participant, setParticipant] = useState<Participant | null>(null)
  const [currentQuestion, setCurrentQuestion] = useState<Question | null>(null)
  const [quizSession, setQuizSession] = useState<QuizSession | null>(null)
  const [selectedAnswer, setSelectedAnswer] = useState<AnswerOption | undefined>()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState<any>(null)

  // 結果取得処理
  const loadResult = useCallback(async () => {
    if (!participant) return

    try {
      const resultData = await apiClient.getParticipantResult(participant.id)
      setResult(resultData)
      setState('result')
    } catch (err: any) {
      console.error('Failed to load result:', err)
      setError('結果の取得に失敗しました')
    }
  }, [participant])

  // WebSocketメッセージハンドラ
  const handleWebSocketMessage = useCallback((message: WebSocketMessage) => {
    console.log('Received WebSocket message:', message)

    switch (message.type) {
      case 'question_switch':
        const questionData = message.data as QuestionSwitchData
        setCurrentQuestion(questionData.question)
        setSelectedAnswer(undefined)
        setState('question')
        break

      case 'voting_end':
        setState('voting_ended')
        break

      case 'session_update':
        const sessionData = message.data as SessionUpdateData
        setQuizSession(sessionData.session)
        
        if (sessionData.session.status === 'finished') {
          loadResult()
        } else if (sessionData.session.status === 'waiting') {
          setState('waiting')
        }
        break

      case 'result_update':
        loadResult()
        break
    }
  }, [loadResult])

  // クイズ参加処理
  const handleJoinQuiz = async (nickname: string) => {
    setLoading(true)
    setError('')

    try {
      const participantData = await apiClient.joinQuiz(nickname, quizId)
      setParticipant(participantData)

      // WebSocket接続
      await wsClient.connect()
      wsClient.subscribe(quizId)
      wsClient.on('*', handleWebSocketMessage)

      // 現在のクイズ状態を取得
      const session = await apiClient.getQuizSession(quizId)
      setQuizSession(session)

      if (session.status === 'question') {
        // 現在の問題を取得
        const question = await apiClient.getCurrentQuestion(quizId)
        setCurrentQuestion(question)
        setState('question')
      } else if (session.status === 'finished') {
        loadResult()
      } else {
        setState('waiting')
      }
    } catch (err: any) {
      setError(err.message || 'クイズへの参加に失敗しました')
    } finally {
      setLoading(false)
    }
  }

  // 回答送信処理
  const handleAnswer = async (option: AnswerOption) => {
    if (!participant || !currentQuestion) return

    const optionIndex = ['A', 'B', 'C', 'D'].indexOf(option)
    setSelectedAnswer(option)

    try {
      await apiClient.submitAnswer(participant.id, currentQuestion.id, optionIndex)
    } catch (err: any) {
      console.error('Failed to submit answer:', err)
      setError('回答の送信に失敗しました')
    }
  }


  // 再開処理
  const handleRestart = () => {
    setParticipant(null)
    setCurrentQuestion(null)
    setQuizSession(null)
    setSelectedAnswer(undefined)
    setResult(null)
    setError('')
    setState('nickname')
    wsClient.disconnect()
  }

  // クリーンアップ
  useEffect(() => {
    return () => {
      wsClient.disconnect()
    }
  }, [])

  // エラー表示
  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md w-full">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className="w-5 h-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-lg font-medium text-red-800 mb-2">
                エラーが発生しました
              </h3>
              <p className="text-red-700 mb-4">
                {error}
              </p>
              <button
                onClick={handleRestart}
                className="btn-primary bg-red-600 hover:bg-red-700 border-red-600 hover:border-red-700"
              >
                最初からやり直す
              </button>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // 状態に応じた画面表示
  switch (state) {
    case 'nickname':
      return <NicknameInput onSubmit={handleJoinQuiz} loading={loading} />

    case 'waiting':
      return (
        <WaitingScreen
          nickname={participant?.nickname || ''}
          currentQuestionNumber={quizSession?.currentQuestionNumber}
          totalQuestions={quizSession?.totalQuestions}
        />
      )

    case 'question':
      if (!currentQuestion || !participant || !quizSession) {
        return <WaitingScreen nickname={participant?.nickname || ''} />
      }
      return (
        <AnswerScreen
          question={currentQuestion}
          currentQuestionNumber={quizSession.currentQuestionNumber}
          totalQuestions={quizSession.totalQuestions}
          onAnswer={handleAnswer}
          selectedAnswer={selectedAnswer}
          isVotingEnded={false}
          nickname={participant.nickname}
        />
      )

    case 'voting_ended':
      if (!currentQuestion || !participant || !quizSession) {
        return <WaitingScreen nickname={participant?.nickname || ''} />
      }
      return (
        <AnswerScreen
          question={currentQuestion}
          currentQuestionNumber={quizSession.currentQuestionNumber}
          totalQuestions={quizSession.totalQuestions}
          onAnswer={handleAnswer}
          selectedAnswer={selectedAnswer}
          isVotingEnded={true}
          nickname={participant.nickname}
        />
      )

    case 'result':
      if (!result || !participant) {
        return <WaitingScreen nickname={participant?.nickname || ''} message="結果を集計中" />
      }
      return (
        <ResultScreen
          nickname={participant.nickname}
          result={result}
          totalParticipants={result.totalParticipants || 1}
          onRestart={handleRestart}
        />
      )

    default:
      return <WaitingScreen nickname={participant?.nickname || ''} />
  }
}