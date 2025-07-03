'use client'

import { useState, useEffect } from 'react'
import { Question, AnswerOption } from '@/types'

interface AnswerScreenProps {
  question: Question
  currentQuestionNumber: number
  totalQuestions: number
  onAnswer: (option: AnswerOption) => void
  selectedAnswer?: AnswerOption
  isVotingEnded: boolean
  nickname: string
}

export default function AnswerScreen({
  question,
  currentQuestionNumber,
  totalQuestions,
  onAnswer,
  selectedAnswer,
  isVotingEnded,
  nickname
}: AnswerScreenProps) {
  const [currentSelected, setCurrentSelected] = useState<AnswerOption | undefined>(selectedAnswer)

  useEffect(() => {
    setCurrentSelected(selectedAnswer)
  }, [selectedAnswer])

  const handleAnswerSelect = (option: AnswerOption) => {
    if (isVotingEnded) return
    
    setCurrentSelected(option)
    onAnswer(option)
  }

  const answerOptions: AnswerOption[] = ['A', 'B', 'C', 'D']

  return (
    <div className="min-h-screen flex flex-col p-4">
      {/* ヘッダー */}
      <div className="bg-white rounded-lg shadow-md p-4 mb-4">
        <div className="flex justify-between items-center mb-2">
          <span className="text-lg font-medium text-gray-600">
            {nickname}
          </span>
          <span className="text-lg font-medium text-blue-600">
            問題 {currentQuestionNumber} / {totalQuestions}
          </span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-3">
          <div 
            className="bg-blue-600 h-3 rounded-full transition-all duration-300"
            style={{ width: `${(currentQuestionNumber / totalQuestions) * 100}%` }}
          />
        </div>
      </div>

      {/* 問題文 */}
      <div className="bg-white rounded-lg shadow-md p-6 mb-6 flex-grow">
        <h1 className="text-2xl font-bold text-gray-900 leading-relaxed">
          {question.text}
        </h1>
      </div>

      {/* 回答ボタン */}
      <div className="space-y-4">
        {answerOptions.map((option, index) => (
          <button
            key={option}
            onClick={() => handleAnswerSelect(option)}
            disabled={isVotingEnded}
            className={`
              btn-answer w-full text-left flex items-center justify-between
              ${currentSelected === option ? 'selected' : ''}
              ${isVotingEnded ? 'opacity-75 cursor-not-allowed' : ''}
            `}
            aria-pressed={currentSelected === option}
            aria-label={`選択肢${option}: ${question.options[index]}`}
          >
            <div className="flex items-center">
              <span className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-gray-200 text-gray-800 font-bold text-xl mr-4">
                {option}
              </span>
              <span className="text-xl">
                {question.options[index]}
              </span>
            </div>
            {currentSelected === option && (
              <div className="w-6 h-6 bg-white rounded-full flex items-center justify-center">
                <div className="w-3 h-3 bg-blue-600 rounded-full" />
              </div>
            )}
          </button>
        ))}
      </div>

      {/* 状態表示 */}
      <div className="mt-6 text-center">
        {isVotingEnded ? (
          <p className="text-lg text-red-600 font-medium">
            投票が終了しました
          </p>
        ) : currentSelected ? (
          <p className="text-lg text-green-600 font-medium">
            {currentSelected} を選択中（変更可能）
          </p>
        ) : (
          <p className="text-lg text-gray-600">
            回答を選択してください
          </p>
        )}
      </div>
    </div>
  )
}