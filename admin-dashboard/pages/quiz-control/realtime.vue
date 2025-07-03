<template>
  <div class="max-w-7xl mx-auto">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-gray-900">リアルタイム結果表示</h1>
      <div class="flex space-x-4">
        <NuxtLink to="/quiz-control" class="btn-secondary">
          制御画面に戻る
        </NuxtLink>
        <button @click="toggleFullscreen" class="btn-primary">
          {{ isFullscreen ? '全画面解除' : '全画面表示' }}
        </button>
      </div>
    </div>

    <div v-if="!session.id" class="text-center py-12">
      <p class="text-gray-500 text-lg">アクティブなクイズセッションがありません</p>
      <NuxtLink to="/quiz-control" class="btn-primary mt-4">
        クイズを開始する
      </NuxtLink>
    </div>

    <div v-else class="space-y-6">
      <!-- セッション情報 -->
      <div class="card">
        <div class="flex justify-between items-center">
          <div>
            <h2 class="text-xl font-semibold text-gray-900">{{ session.title }}</h2>
            <p class="text-gray-600">{{ session.description }}</p>
          </div>
          <div class="text-right">
            <div class="text-sm text-gray-600">問題 {{ session.current_question_number }} / {{ session.total_questions }}</div>
            <div :class="getStatusClass(session.status)" class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium mt-1">
              {{ getStatusLabel(session.status) }}
            </div>
          </div>
        </div>
      </div>

      <!-- 現在の問題 -->
      <div class="card" v-if="currentQuestion">
        <div class="text-center mb-6">
          <h3 class="text-2xl font-bold text-gray-900 mb-2">
            問題 {{ session.current_question_number }}
          </h3>
          <p class="text-lg text-gray-700">{{ currentQuestion.question_text }}</p>
          
          <div v-if="currentQuestion.image_url" class="mt-4 flex justify-center">
            <img :src="currentQuestion.image_url" alt="問題画像" class="max-w-md rounded-lg shadow-sm" />
          </div>
        </div>

        <!-- 選択肢表示 -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
          <div v-for="(option, index) in currentQuestion.options" :key="index" 
               class="p-4 border-2 rounded-lg text-center font-medium text-lg"
               :class="getOptionClass(index)">
            <span class="block text-xl font-bold mb-2">{{ String.fromCharCode(65 + index) }}</span>
            {{ option }}
          </div>
        </div>
      </div>

      <!-- リアルタイムチャート -->
      <RealtimeChart 
        :quiz-id="session.id" 
        :question-id="currentQuestion?.id"
        class="mb-6"
      />

      <!-- 参加者統計 -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div class="card text-center">
          <div class="text-3xl font-bold text-blue-600 mb-2">{{ participantStats.total }}</div>
          <div class="text-gray-600">総参加者数</div>
        </div>
        
        <div class="card text-center">
          <div class="text-3xl font-bold text-green-600 mb-2">{{ participantStats.answered }}</div>
          <div class="text-gray-600">回答済み</div>
        </div>
        
        <div class="card text-center">
          <div class="text-3xl font-bold text-orange-600 mb-2">{{ participantStats.waiting }}</div>
          <div class="text-gray-600">未回答</div>
        </div>
        
        <div class="card text-center">
          <div class="text-3xl font-bold text-purple-600 mb-2">{{ participantStats.rate }}%</div>
          <div class="text-gray-600">回答率</div>
        </div>
      </div>

      <!-- 制限時間表示 -->
      <div class="card" v-if="timeRemaining !== null && session.status === 'active'">
        <div class="text-center">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">残り時間</h3>
          <div class="relative w-32 h-32 mx-auto">
            <svg class="w-32 h-32 transform -rotate-90" viewBox="0 0 120 120">
              <circle
                cx="60"
                cy="60"
                r="54"
                stroke="#e5e7eb"
                stroke-width="12"
                fill="none"
              />
              <circle
                cx="60"
                cy="60"
                r="54"
                :stroke="getTimeColor()"
                stroke-width="12"
                fill="none"
                stroke-linecap="round"
                :stroke-dasharray="circumference"
                :stroke-dashoffset="dashOffset"
                class="transition-all duration-1000"
              />
            </svg>
            <div class="absolute inset-0 flex items-center justify-center">
              <span :class="getTimeTextClass()" class="text-2xl font-bold">{{ timeRemaining }}s</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
definePageMeta({
  middleware: 'auth'
})

const { $fetch } = useNuxtApp()
const config = useRuntimeConfig()

const session = reactive({
  id: null,
  title: '',
  description: '',
  status: 'idle',
  current_question_number: 0,
  total_questions: 0
})

const currentQuestion = ref(null)
const wsConnected = ref(false)
const ws = ref(null)
const isFullscreen = ref(false)
const timeRemaining = ref(null)
const timer = ref(null)

const participantStats = reactive({
  total: 0,
  answered: 0,
  waiting: 0,
  rate: 0
})

const circumference = 2 * Math.PI * 54

const dashOffset = computed(() => {
  if (!currentQuestion.value?.time_limit || timeRemaining.value === null) return circumference
  const progress = timeRemaining.value / currentQuestion.value.time_limit
  return circumference * (1 - progress)
})

const connectWebSocket = () => {
  if (ws.value) return
  
  const wsUrl = `${config.public.wsBase}/ws`
  ws.value = new WebSocket(wsUrl)
  
  ws.value.onopen = () => {
    wsConnected.value = true
    if (session.id) {
      ws.value.send(JSON.stringify({
        type: 'subscribe',
        quiz_id: session.id
      }))
    }
  }
  
  ws.value.onclose = () => {
    wsConnected.value = false
    setTimeout(connectWebSocket, 3000)
  }
  
  ws.value.onmessage = (event) => {
    const data = JSON.parse(event.data)
    handleWebSocketMessage(data)
  }
}

const handleWebSocketMessage = (data) => {
  switch (data.type) {
    case 'answer_status':
      participantStats.total = data.total_participants
      participantStats.answered = data.answered_count
      participantStats.waiting = data.total_participants - data.answered_count
      participantStats.rate = data.total_participants > 0 
        ? Math.round((data.answered_count / data.total_participants) * 100) 
        : 0
      break
    
    case 'question_switch':
      loadCurrentQuestion()
      startTimer()
      break
    
    case 'voting_end':
      stopTimer()
      break
    
    case 'session_update':
      Object.assign(session, data.session)
      break
  }
}

const loadCurrentQuestion = async () => {
  if (!session.id) return
  
  try {
    const response = await $fetch(`/api/quiz/sessions/${session.id}/current-question`, {
      baseURL: config.public.apiBase,
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    currentQuestion.value = response
  } catch (error) {
    console.error('Load current question failed:', error)
  }
}

const loadSession = async () => {
  try {
    const response = await $fetch('/api/quiz/sessions/active', {
      baseURL: config.public.apiBase,
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    
    if (response) {
      Object.assign(session, response)
      await loadCurrentQuestion()
      
      if (ws.value && wsConnected.value) {
        ws.value.send(JSON.stringify({
          type: 'subscribe',
          quiz_id: session.id
        }))
      }
    }
  } catch (error) {
    console.error('Load session failed:', error)
  }
}

const startTimer = () => {
  stopTimer()
  if (!currentQuestion.value?.time_limit) return
  
  timeRemaining.value = currentQuestion.value.time_limit
  
  timer.value = setInterval(() => {
    timeRemaining.value--
    if (timeRemaining.value <= 0) {
      stopTimer()
    }
  }, 1000)
}

const stopTimer = () => {
  if (timer.value) {
    clearInterval(timer.value)
    timer.value = null
  }
  timeRemaining.value = null
}

const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
    isFullscreen.value = true
  } else {
    document.exitFullscreen()
    isFullscreen.value = false
  }
}

const getStatusClass = (status) => {
  switch (status) {
    case 'active': return 'bg-green-100 text-green-800'
    case 'waiting': return 'bg-yellow-100 text-yellow-800'
    case 'completed': return 'bg-blue-100 text-blue-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

const getStatusLabel = (status) => {
  switch (status) {
    case 'active': return '回答受付中'
    case 'waiting': return '待機中'
    case 'completed': return '完了'
    default: return '停止中'
  }
}

const getOptionClass = (index) => {
  const baseClass = 'border-gray-300 text-gray-700'
  // 正解が判明している場合の色分けはここで実装
  return baseClass
}

const getTimeColor = () => {
  if (!currentQuestion.value?.time_limit || timeRemaining.value === null) return '#3B82F6'
  
  const ratio = timeRemaining.value / currentQuestion.value.time_limit
  if (ratio > 0.5) return '#10B981' // green
  if (ratio > 0.25) return '#F59E0B' // orange
  return '#EF4444' // red
}

const getTimeTextClass = () => {
  if (!currentQuestion.value?.time_limit || timeRemaining.value === null) return 'text-blue-600'
  
  const ratio = timeRemaining.value / currentQuestion.value.time_limit
  if (ratio > 0.5) return 'text-green-600'
  if (ratio > 0.25) return 'text-orange-600'
  return 'text-red-600'
}

onMounted(() => {
  connectWebSocket()
  loadSession()
  
  document.addEventListener('fullscreenchange', () => {
    isFullscreen.value = !!document.fullscreenElement
  })
})

onUnmounted(() => {
  if (ws.value) {
    ws.value.close()
  }
  stopTimer()
})
</script>