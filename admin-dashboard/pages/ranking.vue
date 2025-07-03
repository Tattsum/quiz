<template>
  <div class="max-w-7xl mx-auto">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-gray-900">ランキング表示</h1>
      <div class="flex space-x-4">
        <select v-model="selectedSession" @change="loadRanking" class="form-input">
          <option value="">セッションを選択</option>
          <option v-for="sess in sessions" :key="sess.id" :value="sess.id">
            {{ sess.title }} ({{ formatDate(sess.created_at) }})
          </option>
        </select>
        <button @click="refreshRanking" :disabled="!selectedSession || loading" class="btn-secondary">
          更新
        </button>
        <button @click="toggleFullscreen" class="btn-primary">
          {{ isFullscreen ? '全画面解除' : '全画面表示' }}
        </button>
      </div>
    </div>

    <div v-if="!selectedSession" class="text-center py-12">
      <p class="text-gray-500 text-lg">セッションを選択してください</p>
    </div>

    <div v-else-if="loading" class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      <p class="text-gray-500 mt-2">ランキングを読み込み中...</p>
    </div>

    <div v-else class="space-y-6">
      <!-- セッション情報 -->
      <div class="card">
        <div class="flex justify-between items-center">
          <div>
            <h2 class="text-xl font-semibold text-gray-900">{{ sessionInfo.title }}</h2>
            <p class="text-gray-600">{{ sessionInfo.description }}</p>
          </div>
          <div class="text-right">
            <div class="text-sm text-gray-600">参加者: {{ ranking.length }}人</div>
            <div class="text-sm text-gray-600">実施日: {{ formatDate(sessionInfo.created_at) }}</div>
          </div>
        </div>
      </div>

      <!-- 上位3位の表彰台 -->
      <div class="card bg-gradient-to-br from-yellow-50 to-orange-50" v-if="ranking.length >= 3">
        <h3 class="text-xl font-semibold text-gray-900 mb-6 text-center">🏆 上位入賞者 🏆</h3>
        
        <div class="flex justify-center items-end space-x-8">
          <!-- 2位 -->
          <div class="text-center" v-if="ranking[1]">
            <div class="w-20 h-16 bg-silver rounded-t-lg flex items-center justify-center mb-2 mx-auto bg-gray-300">
              <span class="text-white font-bold text-lg">2</span>
            </div>
            <div class="bg-white p-4 rounded-lg shadow-lg min-w-[120px]">
              <div class="text-lg font-bold text-gray-900">{{ ranking[1].participant_name }}</div>
              <div class="text-2xl font-bold text-gray-600 mt-1">{{ ranking[1].total_score }}点</div>
              <div class="text-sm text-gray-500">正解: {{ ranking[1].correct_answers }}問</div>
            </div>
          </div>

          <!-- 1位 -->
          <div class="text-center" v-if="ranking[0]">
            <div class="w-24 h-20 bg-yellow-400 rounded-t-lg flex items-center justify-center mb-2 mx-auto">
              <span class="text-white font-bold text-xl">1</span>
            </div>
            <div class="bg-white p-6 rounded-lg shadow-xl min-w-[140px] border-2 border-yellow-300">
              <div class="text-xl font-bold text-gray-900">{{ ranking[0].participant_name }}</div>
              <div class="text-3xl font-bold text-yellow-600 mt-1">{{ ranking[0].total_score }}点</div>
              <div class="text-sm text-gray-500">正解: {{ ranking[0].correct_answers }}問</div>
              <div class="text-xs text-yellow-600 mt-2">👑 優勝</div>
            </div>
          </div>

          <!-- 3位 -->
          <div class="text-center" v-if="ranking[2]">
            <div class="w-18 h-14 bg-orange-400 rounded-t-lg flex items-center justify-center mb-2 mx-auto">
              <span class="text-white font-bold">3</span>
            </div>
            <div class="bg-white p-3 rounded-lg shadow-lg min-w-[110px]">
              <div class="text-lg font-bold text-gray-900">{{ ranking[2].participant_name }}</div>
              <div class="text-xl font-bold text-orange-600 mt-1">{{ ranking[2].total_score }}点</div>
              <div class="text-sm text-gray-500">正解: {{ ranking[2].correct_answers }}問</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 全体ランキング -->
      <div class="card">
        <h3 class="text-xl font-semibold text-gray-900 mb-4">全体ランキング</h3>
        
        <div class="overflow-x-auto">
          <table class="min-w-full">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">順位</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">参加者名</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">総合得点</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">正解数</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">正解率</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">平均回答時間</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="(participant, index) in ranking" :key="participant.id" 
                  :class="getRankRowClass(index + 1)">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="flex items-center">
                    <span :class="getRankBadgeClass(index + 1)" class="inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-bold">
                      {{ index + 1 }}
                    </span>
                    <span v-if="index < 3" class="ml-2">{{ getRankEmoji(index + 1) }}</span>
                  </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm font-medium text-gray-900">{{ participant.participant_name }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-lg font-bold text-gray-900">{{ participant.total_score }}点</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">{{ participant.correct_answers }}問</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">{{ getAccuracyRate(participant) }}%</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="text-sm text-gray-900">{{ getAverageTime(participant) }}秒</div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 統計情報 -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div class="card text-center">
          <div class="text-2xl font-bold text-blue-600 mb-2">{{ ranking.length }}</div>
          <div class="text-gray-600">総参加者数</div>
        </div>
        
        <div class="card text-center">
          <div class="text-2xl font-bold text-green-600 mb-2">{{ stats.averageScore }}</div>
          <div class="text-gray-600">平均得点</div>
        </div>
        
        <div class="card text-center">
          <div class="text-2xl font-bold text-orange-600 mb-2">{{ stats.topScore }}</div>
          <div class="text-gray-600">最高得点</div>
        </div>
        
        <div class="card text-center">
          <div class="text-2xl font-bold text-purple-600 mb-2">{{ stats.averageAccuracy }}%</div>
          <div class="text-gray-600">平均正解率</div>
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

const sessions = ref([])
const selectedSession = ref('')
const ranking = ref([])
const sessionInfo = ref({})
const loading = ref(false)
const isFullscreen = ref(false)

const stats = computed(() => {
  if (ranking.value.length === 0) return {
    averageScore: 0,
    topScore: 0,
    averageAccuracy: 0
  }

  const scores = ranking.value.map(p => p.total_score)
  const accuracies = ranking.value.map(p => getAccuracyRate(p))

  return {
    averageScore: Math.round(scores.reduce((a, b) => a + b, 0) / scores.length),
    topScore: Math.max(...scores),
    averageAccuracy: Math.round(accuracies.reduce((a, b) => a + b, 0) / accuracies.length)
  }
})

const loadSessions = async () => {
  try {
    const response = await $fetch('/api/quiz/sessions', {
      baseURL: config.public.apiBase,
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    sessions.value = response
  } catch (error) {
    console.error('Sessions load failed:', error)
  }
}

const loadRanking = async () => {
  if (!selectedSession.value) return
  
  loading.value = true
  try {
    const [rankingResponse, sessionResponse] = await Promise.all([
      $fetch(`/api/quiz/sessions/${selectedSession.value}/ranking`, {
        baseURL: config.public.apiBase,
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
        }
      }),
      $fetch(`/api/quiz/sessions/${selectedSession.value}`, {
        baseURL: config.public.apiBase,
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
        }
      })
    ])
    
    ranking.value = rankingResponse
    sessionInfo.value = sessionResponse
  } catch (error) {
    console.error('Ranking load failed:', error)
    alert('ランキングの読み込みに失敗しました。')
  } finally {
    loading.value = false
  }
}

const refreshRanking = () => {
  loadRanking()
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

const getRankRowClass = (rank) => {
  switch (rank) {
    case 1: return 'bg-yellow-50'
    case 2: return 'bg-gray-50'
    case 3: return 'bg-orange-50'
    default: return ''
  }
}

const getRankBadgeClass = (rank) => {
  switch (rank) {
    case 1: return 'bg-yellow-400 text-white'
    case 2: return 'bg-gray-400 text-white'
    case 3: return 'bg-orange-400 text-white'
    default: return 'bg-gray-200 text-gray-700'
  }
}

const getRankEmoji = (rank) => {
  switch (rank) {
    case 1: return '🥇'
    case 2: return '🥈'
    case 3: return '🥉'
    default: return ''
  }
}

const getAccuracyRate = (participant) => {
  if (!participant.total_questions || participant.total_questions === 0) return 0
  return Math.round((participant.correct_answers / participant.total_questions) * 100)
}

const getAverageTime = (participant) => {
  if (!participant.total_answer_time || !participant.total_questions) return 0
  return Math.round(participant.total_answer_time / participant.total_questions)
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  loadSessions()
  
  document.addEventListener('fullscreenchange', () => {
    isFullscreen.value = !!document.fullscreenElement
  })
})
</script>

<style scoped>
.bg-silver {
  background: linear-gradient(135deg, #c0c0c0 0%, #silver 100%);
}
</style>