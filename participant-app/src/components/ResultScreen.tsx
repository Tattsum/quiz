'use client'

import { useEffect, useState } from 'react'

interface ParticipantResult {
  rank: number
  nickname: string
  score: number
  totalQuestions: number
  isCurrentUser: boolean
}

interface ResultScreenProps {
  nickname: string
  result: ParticipantResult
  totalParticipants: number
  onRestart?: () => void
}

export default function ResultScreen({
  nickname,
  result,
  totalParticipants,
  onRestart
}: ResultScreenProps) {
  const [showAnimation, setShowAnimation] = useState(false)

  useEffect(() => {
    const timer = setTimeout(() => {
      setShowAnimation(true)
    }, 500)

    return () => clearTimeout(timer)
  }, [])

  const getRankMessage = (rank: number, total: number) => {
    const percentage = Math.round((rank / total) * 100)
    
    if (rank === 1) return '🏆 優勝！'
    if (rank <= 3) return '🥉 上位入賞！'
    if (percentage <= 25) return '✨ 上位25%！'
    if (percentage <= 50) return '👍 上位50%！'
    return '💪 お疲れ様でした！'
  }

  const getScoreColor = (score: number, total: number) => {
    const percentage = (score / total) * 100
    if (percentage >= 80) return 'text-green-600'
    if (percentage >= 60) return 'text-blue-600'
    if (percentage >= 40) return 'text-yellow-600'
    return 'text-gray-600'
  }

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-md space-y-6">
        {/* 結果ヘッダー */}
        <div className="text-center space-y-2">
          <h1 className="text-3xl font-bold text-gray-900">
            クイズ終了！
          </h1>
          <p className="text-lg text-gray-600">
            お疲れ様でした
          </p>
        </div>

        {/* メイン結果カード */}
        <div className={`bg-white rounded-lg shadow-lg p-6 transform transition-all duration-1000 ${showAnimation ? 'scale-100 opacity-100' : 'scale-95 opacity-0'}`}>
          <div className="text-center space-y-4">
            {/* ニックネーム */}
            <div>
              <p className="text-lg text-gray-600 mb-1">参加者</p>
              <p className="text-2xl font-bold text-blue-600">
                {nickname}
              </p>
            </div>

            {/* 順位 */}
            <div className="py-4">
              <div className="text-4xl mb-2">
                {getRankMessage(result.rank, totalParticipants).split(' ')[0]}
              </div>
              <p className="text-2xl font-bold text-gray-900">
                第 {result.rank} 位
              </p>
              <p className="text-lg text-gray-600">
                {totalParticipants} 人中
              </p>
            </div>

            {/* スコア */}
            <div className="border-t pt-4">
              <p className="text-lg text-gray-600 mb-2">正解数</p>
              <p className={`text-4xl font-bold ${getScoreColor(result.score, result.totalQuestions)}`}>
                {result.score} / {result.totalQuestions}
              </p>
              <p className="text-lg text-gray-600 mt-2">
                正答率: {Math.round((result.score / result.totalQuestions) * 100)}%
              </p>
            </div>
          </div>
        </div>

        {/* 祝福メッセージ */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="text-center">
            <p className="text-xl font-medium text-blue-800">
              {getRankMessage(result.rank, totalParticipants)}
            </p>
            <p className="text-blue-600 mt-1">
              素晴らしい結果です！
            </p>
          </div>
        </div>

        {/* 統計情報 */}
        <div className="bg-gray-50 rounded-lg p-4">
          <h3 className="text-lg font-medium text-gray-900 mb-3">
            クイズ統計
          </h3>
          <div className="grid grid-cols-2 gap-4 text-center">
            <div>
              <p className="text-2xl font-bold text-gray-900">
                {totalParticipants}
              </p>
              <p className="text-sm text-gray-600">
                参加者数
              </p>
            </div>
            <div>
              <p className="text-2xl font-bold text-gray-900">
                {result.totalQuestions}
              </p>
              <p className="text-sm text-gray-600">
                問題数
              </p>
            </div>
          </div>
        </div>

        {/* アクションボタン */}
        {onRestart && (
          <div className="space-y-3">
            <button
              onClick={onRestart}
              className="btn-primary w-full"
            >
              もう一度参加する
            </button>
          </div>
        )}

        {/* フッター */}
        <div className="text-center text-gray-500 text-sm">
          <p>ありがとうございました！</p>
        </div>
      </div>
    </div>
  )
}