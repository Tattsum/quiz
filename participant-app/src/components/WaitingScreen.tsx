'use client'

import { useEffect, useState } from 'react'

interface WaitingScreenProps {
  nickname: string
  currentQuestionNumber?: number
  totalQuestions?: number
  message?: string
}

export default function WaitingScreen({
  nickname,
  currentQuestionNumber,
  totalQuestions,
  message = '次の問題をお待ちください'
}: WaitingScreenProps) {
  const [dots, setDots] = useState('')

  useEffect(() => {
    const interval = setInterval(() => {
      setDots(prev => {
        if (prev === '...') return ''
        return prev + '.'
      })
    }, 500)

    return () => clearInterval(interval)
  }, [])

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4">
      <div className="text-center space-y-8 max-w-md w-full">
        {/* ニックネーム表示 */}
        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-xl font-medium text-gray-600 mb-2">
            参加者
          </h2>
          <p className="text-3xl font-bold text-blue-600">
            {nickname}
          </p>
        </div>

        {/* 進捗表示 */}
        {currentQuestionNumber && totalQuestions && (
          <div className="bg-white rounded-lg shadow-md p-6">
            <div className="flex justify-between items-center mb-4">
              <span className="text-lg font-medium text-gray-600">
                進捗
              </span>
              <span className="text-lg font-bold text-blue-600">
                {currentQuestionNumber} / {totalQuestions}
              </span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-4">
              <div 
                className="bg-blue-600 h-4 rounded-full transition-all duration-300"
                style={{ width: `${(currentQuestionNumber / totalQuestions) * 100}%` }}
              />
            </div>
          </div>
        )}

        {/* 待機メッセージ */}
        <div className="space-y-6">
          <div className="relative">
            <div className="w-20 h-20 mx-auto">
              <div className="w-20 h-20 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
            </div>
          </div>
          
          <div className="space-y-2">
            <h1 className="text-2xl font-bold text-gray-900">
              {message}{dots}
            </h1>
            <p className="text-lg text-gray-600">
              しばらくお待ちください
            </p>
          </div>
        </div>

        {/* 注意事項 */}
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <div className="flex items-start">
            <div className="flex-shrink-0">
              <svg 
                className="w-5 h-5 text-yellow-400 mt-1" 
                fill="currentColor" 
                viewBox="0 0 20 20"
                aria-hidden="true"
              >
                <path 
                  fillRule="evenodd" 
                  d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" 
                  clipRule="evenodd" 
                />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm text-yellow-800">
                この画面を閉じずにお待ちください
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}