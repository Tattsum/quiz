<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <h1 class="text-3xl font-bold text-gray-900">ダッシュボード</h1>
      <div class="flex space-x-4">
        <NuxtLink to="/questions/create" class="btn-primary">
          新しい問題を作成
        </NuxtLink>
        <NuxtLink to="/quiz-control" class="btn-secondary">
          クイズ開始
        </NuxtLink>
      </div>
    </div>

    <!-- 統計情報 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-blue-100 rounded-lg">
            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">総問題数</p>
            <p class="text-2xl font-semibold text-gray-900">{{ stats.totalQuestions }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-green-100 rounded-lg">
            <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">参加者数</p>
            <p class="text-2xl font-semibold text-gray-900">{{ stats.totalParticipants }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-yellow-100 rounded-lg">
            <svg class="w-6 h-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">進行中セッション</p>
            <p class="text-2xl font-semibold text-gray-900">{{ stats.activeSessions }}</p>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="flex items-center">
          <div class="p-2 bg-purple-100 rounded-lg">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
            </svg>
          </div>
          <div class="ml-4">
            <p class="text-sm font-medium text-gray-600">完了セッション</p>
            <p class="text-2xl font-semibold text-gray-900">{{ stats.completedSessions }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 最近の問題 -->
    <div class="card">
      <div class="flex justify-between items-center mb-4">
        <h2 class="text-xl font-semibold text-gray-900">最近の問題</h2>
        <NuxtLink to="/questions" class="text-blue-600 hover:text-blue-800 text-sm font-medium">
          すべて表示 →
        </NuxtLink>
      </div>
      
      <div class="overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">問題</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">カテゴリ</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">作成日</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="question in recentQuestions" :key="question.id">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ question.question_text }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ question.category }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ formatDate(question.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <NuxtLink :to="`/questions/${question.id}/edit`" class="text-blue-600 hover:text-blue-900 mr-3">
                  編集
                </NuxtLink>
                <button @click="deleteQuestion(question.id)" class="text-red-600 hover:text-red-900">
                  削除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
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

const stats = reactive({
  totalQuestions: 0,
  totalParticipants: 0,
  activeSessions: 0,
  completedSessions: 0
})

const recentQuestions = ref([])

const fetchStats = async () => {
  try {
    const response = await $fetch('/api/admin/stats', {
      baseURL: config.public.apiBase,
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    Object.assign(stats, response)
  } catch (error) {
    console.error('Stats fetch failed:', error)
  }
}

const fetchRecentQuestions = async () => {
  try {
    const response = await $fetch('/api/questions?limit=5', {
      baseURL: config.public.apiBase,
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    recentQuestions.value = response
  } catch (error) {
    console.error('Recent questions fetch failed:', error)
  }
}

const deleteQuestion = async (id) => {
  if (!confirm('この問題を削除しますか？')) return
  
  try {
    await $fetch(`/api/questions/${id}`, {
      baseURL: config.public.apiBase,
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    await fetchRecentQuestions()
    await fetchStats()
  } catch (error) {
    console.error('Delete failed:', error)
  }
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString('ja-JP')
}

onMounted(() => {
  fetchStats()
  fetchRecentQuestions()
})
</script>