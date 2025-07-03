<template>
  <div class="bg-white rounded-lg shadow-lg p-6">
    <div class="flex justify-between items-center mb-4">
      <h3 class="text-lg font-semibold text-gray-900">リアルタイム回答状況</h3>
      <div class="flex items-center space-x-2">
        <div :class="wsConnected ? 'bg-green-400' : 'bg-red-400'" class="w-2 h-2 rounded-full"></div>
        <span class="text-sm text-gray-600">
          {{ wsConnected ? '接続中' : '切断' }}
        </span>
      </div>
    </div>

    <div v-if="!hasData" class="text-center py-8 text-gray-500">
      回答データがありません
    </div>

    <div v-else class="space-y-6">
      <!-- 全体統計 -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="text-center">
          <div class="text-2xl font-bold text-blue-600">{{ stats.totalParticipants }}</div>
          <div class="text-sm text-gray-600">総参加者</div>
        </div>
        <div class="text-center">
          <div class="text-2xl font-bold text-green-600">{{ stats.answeredCount }}</div>
          <div class="text-sm text-gray-600">回答済み</div>
        </div>
        <div class="text-center">
          <div class="text-2xl font-bold text-orange-600">{{ stats.totalParticipants - stats.answeredCount }}</div>
          <div class="text-sm text-gray-600">未回答</div>
        </div>
        <div class="text-center">
          <div class="text-2xl font-bold text-purple-600">{{ answerRate }}%</div>
          <div class="text-sm text-gray-600">回答率</div>
        </div>
      </div>

      <!-- 円グラフ -->
      <div class="flex justify-center">
        <div class="relative w-80 h-80">
          <canvas ref="chartCanvas" width="320" height="320"></canvas>
          <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
            <div class="text-center">
              <div class="text-3xl font-bold text-gray-900">{{ stats.answeredCount }}</div>
              <div class="text-sm text-gray-600">/ {{ stats.totalParticipants }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 選択肢別詳細 -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div v-for="(count, index) in stats.answerCounts" :key="index" 
             class="flex items-center justify-between p-3 border rounded-lg">
          <div class="flex items-center space-x-3">
            <div 
              :style="`background-color: ${getOptionColor(index)}`"
              class="w-4 h-4 rounded-full"
            ></div>
            <span class="font-medium">選択肢 {{ String.fromCharCode(65 + index) }}</span>
          </div>
          <div class="text-right">
            <div class="font-bold">{{ count }}人</div>
            <div class="text-sm text-gray-600">{{ getPercentage(count) }}%</div>
          </div>
        </div>
      </div>

      <!-- 時系列グラフ（簡易版） -->
      <div class="border-t pt-4">
        <h4 class="text-md font-medium text-gray-900 mb-2">回答推移</h4>
        <div class="h-24 bg-gray-50 rounded flex items-end justify-between px-2 space-x-1">
          <div v-for="(point, index) in timeSeriesData" :key="index" 
               class="bg-blue-500 rounded-t min-w-[8px]"
               :style="`height: ${(point / stats.totalParticipants) * 100}%`"
               :title="`${point}人が回答`"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Chart, ArcElement, CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend } from 'chart.js'

Chart.register(ArcElement, CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

const props = defineProps({
  quizId: {
    type: [String, Number],
    default: null
  },
  questionId: {
    type: [String, Number],
    default: null
  }
})

const config = useRuntimeConfig()
const chartCanvas = ref(null)

const wsConnected = ref(false)
const ws = ref(null)
const chart = ref(null)

const stats = reactive({
  totalParticipants: 0,
  answeredCount: 0,
  answerCounts: [0, 0, 0, 0]
})

const timeSeriesData = ref([])

const hasData = computed(() => {
  return stats.totalParticipants > 0
})

const answerRate = computed(() => {
  if (stats.totalParticipants === 0) return 0
  return Math.round((stats.answeredCount / stats.totalParticipants) * 100)
})

const chartData = computed(() => {
  const labels = stats.answerCounts.map((_, index) => `選択肢 ${String.fromCharCode(65 + index)}`)
  const data = stats.answerCounts
  const backgroundColor = [
    '#3B82F6', // blue
    '#10B981', // green
    '#F59E0B', // orange
    '#EF4444'  // red
  ]

  return {
    labels,
    datasets: [{
      data,
      backgroundColor,
      borderWidth: 2,
      borderColor: '#ffffff'
    }]
  }
})

const connectWebSocket = () => {
  if (ws.value || !props.quizId) return
  
  const wsUrl = `${config.public.wsBase}/ws`
  ws.value = new WebSocket(wsUrl)
  
  ws.value.onopen = () => {
    wsConnected.value = true
    ws.value.send(JSON.stringify({
      type: 'subscribe',
      quiz_id: props.quizId
    }))
  }
  
  ws.value.onclose = () => {
    wsConnected.value = false
    setTimeout(connectWebSocket, 3000)
  }
  
  ws.value.onmessage = (event) => {
    const data = JSON.parse(event.data)
    handleWebSocketMessage(data)
  }
  
  ws.value.onerror = (error) => {
    console.error('WebSocket error:', error)
  }
}

const handleWebSocketMessage = (data) => {
  if (data.type === 'answer_status' && data.question_id === props.questionId) {
    stats.totalParticipants = data.total_participants
    stats.answeredCount = data.answered_count
    stats.answerCounts = data.answer_counts || [0, 0, 0, 0]
    
    // 時系列データを更新（最大20ポイント保持）
    timeSeriesData.value.push(data.answered_count)
    if (timeSeriesData.value.length > 20) {
      timeSeriesData.value.shift()
    }
    
    updateChart()
  }
}

const createChart = () => {
  if (!chartCanvas.value) return
  
  const ctx = chartCanvas.value.getContext('2d')
  
  chart.value = new Chart(ctx, {
    type: 'doughnut',
    data: chartData.value,
    options: {
      responsive: false,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          callbacks: {
            label: function(context) {
              const total = context.dataset.data.reduce((a, b) => a + b, 0)
              const percentage = total > 0 ? Math.round((context.raw / total) * 100) : 0
              return `${context.label}: ${context.raw}人 (${percentage}%)`
            }
          }
        }
      },
      cutout: '60%'
    }
  })
}

const updateChart = () => {
  if (!chart.value) return
  
  chart.value.data = chartData.value
  chart.value.update('none') // アニメーションなしで更新
}

const getOptionColor = (index) => {
  const colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444']
  return colors[index] || '#6B7280'
}

const getPercentage = (count) => {
  if (stats.answeredCount === 0) return 0
  return Math.round((count / stats.answeredCount) * 100)
}

const disconnectWebSocket = () => {
  if (ws.value) {
    ws.value.close()
    ws.value = null
  }
  wsConnected.value = false
}

// Props変更時の再接続
watch([() => props.quizId, () => props.questionId], () => {
  disconnectWebSocket()
  // 統計をリセット
  Object.assign(stats, {
    totalParticipants: 0,
    answeredCount: 0,
    answerCounts: [0, 0, 0, 0]
  })
  timeSeriesData.value = []
  
  if (props.quizId) {
    nextTick(() => {
      connectWebSocket()
    })
  }
})

onMounted(() => {
  nextTick(() => {
    createChart()
    if (props.quizId) {
      connectWebSocket()
    }
  })
})

onUnmounted(() => {
  disconnectWebSocket()
  if (chart.value) {
    chart.value.destroy()
  }
})
</script>