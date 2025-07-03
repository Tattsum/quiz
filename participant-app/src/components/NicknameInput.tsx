'use client'

import { useState } from 'react'

interface NicknameInputProps {
  onSubmit: (nickname: string) => void
  loading?: boolean
}

export default function NicknameInput({ onSubmit, loading = false }: NicknameInputProps) {
  const [nickname, setNickname] = useState('')
  const [error, setError] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!nickname.trim()) {
      setError('ニックネームを入力してください')
      return
    }
    
    if (nickname.trim().length < 2) {
      setError('ニックネームは2文字以上で入力してください')
      return
    }
    
    if (nickname.trim().length > 20) {
      setError('ニックネームは20文字以下で入力してください')
      return
    }
    
    setError('')
    onSubmit(nickname.trim())
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            クイズに参加
          </h1>
          <p className="text-xl text-gray-600">
            ニックネームを入力してください
          </p>
        </div>
        
        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label 
              htmlFor="nickname" 
              className="block text-lg font-medium text-gray-700 mb-2"
            >
              ニックネーム
            </label>
            <input
              id="nickname"
              type="text"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              className="input-primary w-full"
              placeholder="例: たろう"
              maxLength={20}
              disabled={loading}
              autoFocus
              aria-describedby={error ? "nickname-error" : undefined}
            />
            {error && (
              <p 
                id="nickname-error" 
                className="mt-2 text-red-600 text-lg"
                role="alert"
              >
                {error}
              </p>
            )}
          </div>
          
          <button
            type="submit"
            disabled={loading || !nickname.trim()}
            className="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? '参加中...' : 'クイズに参加する'}
          </button>
        </form>
        
        <div className="text-center text-gray-500 text-sm">
          <p>参加者は最大70名まで</p>
        </div>
      </div>
    </div>
  )
}