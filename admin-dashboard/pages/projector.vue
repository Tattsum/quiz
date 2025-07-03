<template>
  <div class="projector-screen">
    <div v-if="!showResult" class="question-view">
      <!-- ヘッダー：問題番号 -->
      <div class="header">
        <h1 class="question-number">問題 {{ currentQuestion?.questionNumber || 1 }} / {{ totalQuestions || 10 }}</h1>
        <div v-if="timeLeft > 0" class="timer">{{ formatTime(timeLeft) }}</div>
      </div>

      <!-- 問題文 -->
      <div class="question-section">
        <h2 class="question-text">{{ currentQuestion?.questionText || 'クイズの問題文がここに表示されます' }}</h2>
        
        <!-- 画像表示エリア -->
        <div v-if="currentQuestion?.imageUrl" class="image-section">
          <img :src="currentQuestion.imageUrl" :alt="currentQuestion.questionText" class="question-image" />
        </div>
      </div>

      <!-- 選択肢 -->
      <div class="choices-section">
        <div class="choices-grid">
          <div 
            v-for="(choice, index) in choices" 
            :key="index"
            class="choice-item"
            :class="`choice-${choice.label.toLowerCase()}`"
          >
            <div class="choice-label">{{ choice.label }}</div>
            <div class="choice-text">{{ choice.text }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 結果表示画面 -->
    <div v-else class="result-view">
      <div class="result-header">
        <h1 class="result-title">正解発表</h1>
        <div class="question-number">問題 {{ currentQuestion?.questionNumber || 1 }}</div>
      </div>

      <div class="correct-answer">
        <h2 class="correct-label">正解</h2>
        <div class="correct-choice" :class="`choice-${correctAnswer?.toLowerCase()}`">
          <div class="choice-label">{{ correctAnswer }}</div>
          <div class="choice-text">{{ getChoiceText(correctAnswer) }}</div>
        </div>
      </div>

      <div class="stats-section">
        <div class="accuracy-rate">
          <h3>正答率</h3>
          <div class="percentage">{{ accuracyRate }}%</div>
        </div>
        
        <div class="answer-distribution">
          <h3>回答分布</h3>
          <div class="distribution-grid">
            <div 
              v-for="(choice, index) in choices" 
              :key="index"
              class="distribution-item"
              :class="{ 'correct': choice.label === correctAnswer }"
            >
              <div class="choice-label">{{ choice.label }}</div>
              <div class="vote-count">{{ getVoteCount(choice.label) }}人</div>
              <div class="vote-percentage">{{ getVotePercentage(choice.label) }}%</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'

// プロジェクター専用レイアウトを設定
definePageMeta({
  layout: 'projector'
})

// リアクティブデータ
const currentQuestion = ref(null)
const totalQuestions = ref(10)
const timeLeft = ref(60)
const showResult = ref(false)
const correctAnswer = ref('A')
const accuracyRate = ref(75)
const answerStats = ref({
  A: 15,
  B: 3,
  C: 2,
  D: 0
})

// WebSocket接続
let ws = null
let timer = null

// 選択肢データ
const choices = computed(() => {
  if (!currentQuestion.value) {
    return [
      { label: 'A', text: '選択肢Aの内容がここに表示されます' },
      { label: 'B', text: '選択肢Bの内容がここに表示されます' },
      { label: 'C', text: '選択肢Cの内容がここに表示されます' },
      { label: 'D', text: '選択肢Dの内容がここに表示されます' }
    ]
  }
  
  return [
    { label: 'A', text: currentQuestion.value.choiceA },
    { label: 'B', text: currentQuestion.value.choiceB },
    { label: 'C', text: currentQuestion.value.choiceC },
    { label: 'D', text: currentQuestion.value.choiceD }
  ]
})

// 時間フォーマット
const formatTime = (seconds) => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

// 選択肢のテキストを取得
const getChoiceText = (label) => {
  const choice = choices.value.find(c => c.label === label)
  return choice ? choice.text : ''
}

// 投票数を取得
const getVoteCount = (label) => {
  return answerStats.value[label] || 0
}

// 投票割合を取得
const getVotePercentage = (label) => {
  const total = Object.values(answerStats.value).reduce((sum, count) => sum + count, 0)
  if (total === 0) return 0
  return Math.round((answerStats.value[label] || 0) / total * 100)
}

// WebSocket接続
const connectWebSocket = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.hostname}:8080/ws`
  
  ws = new WebSocket(wsUrl)
  
  ws.onopen = () => {
    console.log('WebSocket connected')
    // クイズIDを購読（実際のクイズIDに置き換える）
    ws.send(JSON.stringify({
      type: 'subscribe',
      quizId: '1' // 実際のクイズIDに置き換える
    }))
  }
  
  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    handleWebSocketMessage(data)
  }
  
  ws.onclose = () => {
    console.log('WebSocket disconnected')
    // 再接続を試行
    setTimeout(connectWebSocket, 3000)
  }
}

// WebSocketメッセージハンドラー
const handleWebSocketMessage = (data) => {
  switch (data.type) {
    case 'question_switch':
      // 新しい問題に切り替え
      loadQuestion(data.questionNumber)
      showResult.value = false
      startTimer(data.timeLimit || 60)
      break
      
    case 'voting_end':
      // 投票終了、結果表示
      showResult.value = true
      stopTimer()
      loadResults(data.questionId)
      break
      
    case 'answer_status':
      // 回答状況更新
      if (data.answerCounts) {
        answerStats.value = data.answerCounts
      }
      break
  }
}

// 問題データを読み込み
const loadQuestion = async (questionNumber) => {
  try {
    // 実際のAPIエンドポイントに置き換える
    const response = await fetch(`/api/questions/${questionNumber}`)
    if (response.ok) {
      currentQuestion.value = await response.json()
    }
  } catch (error) {
    console.error('Failed to load question:', error)
  }
}

// 結果データを読み込み
const loadResults = async (questionId) => {
  try {
    // 実際のAPIエンドポイントに置き換える
    const response = await fetch(`/api/results/${questionId}`)
    if (response.ok) {
      const results = await response.json()
      correctAnswer.value = results.correctAnswer
      accuracyRate.value = results.accuracyRate
      answerStats.value = results.answerStats
    }
  } catch (error) {
    console.error('Failed to load results:', error)
  }
}

// タイマー開始
const startTimer = (duration) => {
  timeLeft.value = duration
  timer = setInterval(() => {
    if (timeLeft.value > 0) {
      timeLeft.value--
    } else {
      stopTimer()
    }
  }, 1000)
}

// タイマー停止
const stopTimer = () => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

// コンポーネントマウント時
onMounted(() => {
  connectWebSocket()
  // 初期問題を読み込み
  loadQuestion(1)
})

// コンポーネントアンマウント時
onUnmounted(() => {
  stopTimer()
  if (ws) {
    ws.close()
  }
})
</script>

<style scoped>
.projector-screen {
  width: 100vw;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-family: 'Hiragino Kaku Gothic ProN', 'Yu Gothic', sans-serif;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 問題表示画面 */
.question-view {
  height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 2rem;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  background: rgba(255, 255, 255, 0.1);
  padding: 1.5rem 2rem;
  border-radius: 20px;
  backdrop-filter: blur(10px);
}

.question-number {
  font-size: 3rem;
  font-weight: bold;
  margin: 0;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
}

.timer {
  font-size: 4rem;
  font-weight: bold;
  color: #ffeb3b;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.question-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  text-align: center;
  margin-bottom: 2rem;
}

.question-text {
  font-size: 3.5rem;
  line-height: 1.4;
  margin-bottom: 2rem;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
  background: rgba(255, 255, 255, 0.1);
  padding: 2rem;
  border-radius: 20px;
  backdrop-filter: blur(10px);
}

.image-section {
  margin: 2rem 0;
}

.question-image {
  max-width: 60%;
  max-height: 300px;
  border-radius: 15px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.choices-section {
  margin-top: auto;
}

.choices-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.choice-item {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 20px;
  padding: 2rem;
  text-align: center;
  backdrop-filter: blur(10px);
  border: 3px solid transparent;
  transition: all 0.3s ease;
}

.choice-item.choice-a {
  border-color: #ff6b6b;
}

.choice-item.choice-b {
  border-color: #4ecdc4;
}

.choice-item.choice-c {
  border-color: #45b7d1;
}

.choice-item.choice-d {
  border-color: #96ceb4;
}

.choice-label {
  font-size: 4rem;
  font-weight: bold;
  margin-bottom: 1rem;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
}

.choice-text {
  font-size: 2rem;
  line-height: 1.3;
}

/* 結果表示画面 */
.result-view {
  height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 2rem;
}

.result-header {
  text-align: center;
  margin-bottom: 2rem;
  background: rgba(255, 255, 255, 0.1);
  padding: 1.5rem;
  border-radius: 20px;
  backdrop-filter: blur(10px);
}

.result-title {
  font-size: 4rem;
  font-weight: bold;
  margin-bottom: 1rem;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
}

.correct-answer {
  text-align: center;
  margin-bottom: 3rem;
}

.correct-label {
  font-size: 3rem;
  margin-bottom: 1rem;
  color: #4caf50;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
}

.correct-choice {
  background: rgba(76, 175, 80, 0.2);
  border: 4px solid #4caf50;
  border-radius: 20px;
  padding: 2rem;
  max-width: 600px;
  margin: 0 auto;
  backdrop-filter: blur(10px);
}

.stats-section {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 3rem;
  flex: 1;
}

.accuracy-rate {
  text-align: center;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 2rem;
  backdrop-filter: blur(10px);
}

.accuracy-rate h3 {
  font-size: 2.5rem;
  margin-bottom: 1rem;
}

.percentage {
  font-size: 6rem;
  font-weight: bold;
  color: #4caf50;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
}

.answer-distribution {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 2rem;
  backdrop-filter: blur(10px);
}

.answer-distribution h3 {
  font-size: 2.5rem;
  margin-bottom: 2rem;
  text-align: center;
}

.distribution-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
}

.distribution-item {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 15px;
  padding: 1.5rem;
  text-align: center;
  border: 2px solid rgba(255, 255, 255, 0.3);
}

.distribution-item.correct {
  background: rgba(76, 175, 80, 0.2);
  border-color: #4caf50;
}

.distribution-item .choice-label {
  font-size: 2.5rem;
  font-weight: bold;
  margin-bottom: 0.5rem;
}

.vote-count {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.vote-percentage {
  font-size: 1.5rem;
  opacity: 0.8;
}

/* レスポンシブ対応 */
@media (max-width: 1200px) {
  .question-text {
    font-size: 2.5rem;
  }
  
  .choice-text {
    font-size: 1.5rem;
  }
  
  .stats-section {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .choices-grid {
    grid-template-columns: 1fr;
  }
  
  .distribution-grid {
    grid-template-columns: 1fr;
  }
}
</style>