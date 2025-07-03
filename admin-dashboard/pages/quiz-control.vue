<template>
  <div class="max-w-6xl mx-auto">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-gray-900">クイズ制御</h1>
      <div class="flex space-x-4">
        <span :class="getStatusClass(session.status)" class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium">
          {{ getStatusLabel(session.status) }}
        </span>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- メイン制御パネル -->
      <div class="lg:col-span-2 space-y-6">
        <!-- セッション作成/選択 -->
        <div class="card" v-if="!session.id">
          <h2 class="text-xl font-semibold text-gray-900 mb-4">新しいクイズセッションを開始</h2>
          
          <div class="space-y-4">
            <div>
              <label for="title" class="block text-sm font-medium text-gray-700 mb-2">
                セッションタイトル
              </label>
              <input
                id="title"
                v-model="newSession.title"
                type="text"
                required
                class="form-input"
                placeholder="例: 第1回 一般知識クイズ大会"
              />
            </div>
            
            <div>
              <label for="description" class="block text-sm font-medium text-gray-700 mb-2">
                説明
              </label>
              <textarea
                id="description"
                v-model="newSession.description"
                rows="3"
                class="form-textarea"
                placeholder="クイズセッションの説明を入力してください"
              ></textarea>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label for="max_participants" class="block text-sm font-medium text-gray-700 mb-2">
                  最大参加者数
                </label>
                <input
                  id="max_participants"
                  v-model.number="newSession.max_participants"
                  type="number"
                  min="1"
                  max="70"
                  class="form-input"
                  placeholder="70"
                />
              </div>
              
              <div>
                <label for="question_count" class="block text-sm font-medium text-gray-700 mb-2">
                  問題数
                </label>
                <input
                  id="question_count"
                  v-model.number="newSession.question_count"
                  type="number"
                  min="1"
                  max="50"
                  class="form-input"
                  placeholder="10"
                />
              </div>
            </div>
            
            <button @click="createSession" :disabled="loading" class="btn-primary w-full">
              {{ loading ? 'セッション作成中...' : 'セッションを作成して開始' }}
            </button>
          </div>
        </div>

        <!-- アクティブセッション制御 -->
        <div class="card" v-if="session.id">
          <div class="flex justify-between items-start mb-6">
            <div>
              <h2 class="text-xl font-semibold text-gray-900">{{ session.title }}</h2>
              <p class="text-gray-600 mt-1">{{ session.description }}</p>
            </div>
            <button @click="endSession" class="btn-danger">
              セッション終了
            </button>
          </div>

          <!-- 進行状況 -->
          <div class="mb-6">
            <div class="flex justify-between text-sm text-gray-600 mb-2">
              <span>進行状況</span>
              <span>{{ session.current_question_number || 0 }} / {{ session.total_questions || 0 }}</span>
            </div>
            <div class="w-full bg-gray-200 rounded-full h-2">
              <div 
                class="bg-blue-600 h-2 rounded-full transition-all duration-300"
                :style="`width: ${progressPercentage}%`"
              ></div>
            </div>
          </div>

          <!-- 現在の問題 -->
          <div class="bg-gray-50 rounded-lg p-4 mb-6" v-if="currentQuestion">
            <h3 class="text-lg font-medium text-gray-900 mb-2">
              問題 {{ session.current_question_number }}: {{ currentQuestion.question_text }}
            </h3>
            <div class="grid grid-cols-2 gap-2 mt-4">
              <div v-for="(option, index) in currentQuestion.options" :key="index" 
                   :class="['p-2 rounded text-sm', index === currentQuestion.correct_answer ? 'bg-green-100 text-green-800' : 'bg-white']">
                {{ String.fromCharCode(65 + index) }}. {{ option }}
              </div>
            </div>
          </div>

          <!-- 制御ボタン -->
          <div class="flex space-x-4">
            <button 
              @click="startQuestion"
              :disabled="session.status === 'active' || loading"
              class="btn-primary flex-1"
            >
              {{ session.status === 'waiting' ? '問題開始' : '次の問題へ' }}
            </button>
            
            <button 
              @click="endVoting"
              :disabled="session.status !== 'active' || loading"
              class="btn-secondary flex-1"
            >
              投票終了
            </button>
            
            <button 
              @click="showResults"
              :disabled="session.status === 'active' || loading"
              class="btn-secondary flex-1"
            >
              結果表示
            </button>
          </div>
        </div>
      </div>

      <!-- サイドパネル -->
      <div class="space-y-6">
        <!-- 参加者情報 -->
        <div class="card">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">参加者情報</h3>
          
          <div class="space-y-3">
            <div class="flex justify-between">
              <span class="text-gray-600">現在の参加者</span>
              <span class="font-semibold">{{ stats.current_participants || 0 }}人</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">最大参加者数</span>
              <span class="font-semibold">{{ session.max_participants || 0 }}人</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">回答済み</span>
              <span class="font-semibold">{{ stats.answered_count || 0 }}人</span>
            </div>
          </div>

          <div class="mt-4 p-3 bg-blue-50 rounded-lg">
            <p class="text-sm text-blue-800 font-medium">参加URL</p>
            <code class="text-xs text-blue-600 break-all">
              {{ participantUrl }}
            </code>
            <button @click="copyUrl" class="btn-secondary w-full mt-2 text-xs">
              URLをコピー
            </button>
          </div>
        </div>

        <!-- WebSocket接続状況 -->
        <div class="card">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">接続状況</h3>
          
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-gray-600">WebSocket</span>
              <span :class="wsConnected ? 'text-green-600' : 'text-red-600'" class="font-semibold">
                {{ wsConnected ? '接続中' : '切断' }}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">リスナー数</span>
              <span class="font-semibold">{{ stats.listeners || 0 }}</span>
            </div>
          </div>
        </div>

        <!-- クイック操作 -->
        <div class="card">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">クイック操作</h3>
          
          <div class="space-y-2">
            <button @click="refreshStats" class="btn-secondary w-full text-sm">
              統計を更新
            </button>
            <button @click="sendHeartbeat" class="btn-secondary w-full text-sm">
              接続確認
            </button>
            <NuxtLink to="/questions" class="btn-secondary w-full text-sm text-center block">
              問題管理
            </NuxtLink>
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
  total_questions: 0,
  max_participants: 70
})

const newSession = reactive({
  title: '',
  description: '',
  max_participants: 70,
  question_count: 10
})

const currentQuestion = ref(null)
const loading = ref(false)
const wsConnected = ref(false)
const ws = ref(null)

const stats = reactive({
  current_participants: 0,
  answered_count: 0,
  listeners: 0
})

const progressPercentage = computed(() => {
  if (!session.total_questions) return 0
  return (session.current_question_number / session.total_questions) * 100
})

const participantUrl = computed(() => {
  if (!session.id) return ''
  return `${window.location.origin}/quiz/${session.id}`
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
      stats.answered_count = data.answered_count
      stats.current_participants = data.total_participants
      break
    case 'session_update':
      Object.assign(session, data.session)
      break
  }
}

const createSession = async () => {
  loading.value = true
  try {
    const response = await $fetch('/api/quiz/sessions', {
      baseURL: config.public.apiBase,
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      },
      body: newSession
    })
    
    Object.assign(session, response)
    if (ws.value && wsConnected.value) {
      ws.value.send(JSON.stringify({
        type: 'subscribe',
        quiz_id: session.id
      }))
    }
  } catch (error) {
    console.error('Session creation failed:', error)
    alert('セッションの作成に失敗しました。')
  } finally {
    loading.value = false
  }
}

const startQuestion = async () => {
  loading.value = true
  try {
    const response = await $fetch(`/api/quiz/sessions/${session.id}/start-question`, {
      baseURL: config.public.apiBase,
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    
    currentQuestion.value = response.question
    Object.assign(session, response.session)
  } catch (error) {
    console.error('Start question failed:', error)
    alert('問題の開始に失敗しました。')
  } finally {
    loading.value = false
  }
}

const endVoting = async () => {
  loading.value = true
  try {
    await $fetch(`/api/quiz/sessions/${session.id}/end-voting`, {
      baseURL: config.public.apiBase,
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    
    session.status = 'waiting'
  } catch (error) {
    console.error('End voting failed:', error)
    alert('投票終了に失敗しました。')
  } finally {
    loading.value = false
  }
}

const showResults = async () => {
  loading.value = true
  try {
    await $fetch(`/api/quiz/sessions/${session.id}/show-results`, {
      baseURL: config.public.apiBase,
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
  } catch (error) {
    console.error('Show results failed:', error)
    alert('結果表示に失敗しました。')
  } finally {
    loading.value = false
  }
}

const endSession = async () => {
  if (!confirm('セッションを終了しますか？')) return
  
  loading.value = true
  try {
    await $fetch(`/api/quiz/sessions/${session.id}/end`, {
      baseURL: config.public.apiBase,
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    
    Object.assign(session, {
      id: null,
      title: '',
      description: '',
      status: 'idle',
      current_question_number: 0,
      total_questions: 0
    })
    currentQuestion.value = null
  } catch (error) {
    console.error('End session failed:', error)
    alert('セッション終了に失敗しました。')
  } finally {
    loading.value = false
  }
}

const refreshStats = async () => {
  if (!session.id) return
  
  try {
    const response = await $fetch(`/api/quiz/sessions/${session.id}/stats`, {
      baseURL: config.public.apiBase,
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    Object.assign(stats, response)
  } catch (error) {
    console.error('Stats refresh failed:', error)
  }
}

const sendHeartbeat = () => {
  if (ws.value && wsConnected.value) {
    ws.value.send(JSON.stringify({ type: 'heartbeat' }))
  }
}

const copyUrl = async () => {
  try {
    await navigator.clipboard.writeText(participantUrl.value)
    alert('URLをコピーしました。')
  } catch (error) {
    console.error('Copy failed:', error)
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
    case 'active': return '進行中'
    case 'waiting': return '待機中'
    case 'completed': return '完了'
    default: return '停止中'
  }
}

onMounted(() => {
  connectWebSocket()
})

onUnmounted(() => {
  if (ws.value) {
    ws.value.close()
  }
})
</script>