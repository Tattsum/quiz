<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-gray-900">問題管理</h1>
      <NuxtLink to="/questions/create" class="btn-primary">
        新しい問題を作成
      </NuxtLink>
    </div>

    <!-- 検索・フィルター -->
    <div class="card mb-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div>
          <label for="search" class="block text-sm font-medium text-gray-700 mb-2">
            検索
          </label>
          <input
            id="search"
            v-model="filters.search"
            type="text"
            class="form-input"
            placeholder="問題文で検索"
            @input="debouncedSearch"
          />
        </div>
        
        <div>
          <label for="category" class="block text-sm font-medium text-gray-700 mb-2">
            カテゴリ
          </label>
          <select id="category" v-model="filters.category" @change="fetchQuestions" class="form-input">
            <option value="">すべて</option>
            <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
          </select>
        </div>
        
        <div>
          <label for="difficulty" class="block text-sm font-medium text-gray-700 mb-2">
            難易度
          </label>
          <select id="difficulty" v-model="filters.difficulty" @change="fetchQuestions" class="form-input">
            <option value="">すべて</option>
            <option value="easy">易しい</option>
            <option value="medium">普通</option>
            <option value="hard">難しい</option>
          </select>
        </div>
        
        <div class="flex items-end">
          <button @click="resetFilters" class="btn-secondary w-full">
            リセット
          </button>
        </div>
      </div>
    </div>

    <!-- 問題一覧 -->
    <div class="card">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                問題
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                カテゴリ
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                難易度
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                配点
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                作成日
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                操作
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="question in questions" :key="question.id" class="hover:bg-gray-50">
              <td class="px-6 py-4">
                <div class="flex items-center">
                  <div class="flex-shrink-0 w-10 h-10 mr-4" v-if="question.image_url">
                    <img :src="question.image_url" alt="" class="w-10 h-10 rounded object-cover" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-medium text-gray-900 truncate">
                      {{ question.question_text }}
                    </p>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  {{ question.category }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getDifficultyClass(question.difficulty)" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
                  {{ getDifficultyLabel(question.difficulty) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ question.points }}pt
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ formatDate(question.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <NuxtLink :to="`/questions/${question.id}`" class="text-blue-600 hover:text-blue-900">
                    表示
                  </NuxtLink>
                  <NuxtLink :to="`/questions/${question.id}/edit`" class="text-green-600 hover:text-green-900">
                    編集
                  </NuxtLink>
                  <button @click="deleteQuestion(question.id)" class="text-red-600 hover:text-red-900">
                    削除
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- ページネーション -->
      <div class="px-6 py-3 flex items-center justify-between border-t border-gray-200" v-if="pagination.total > 0">
        <div class="flex-1 flex justify-between sm:hidden">
          <button
            @click="changePage(pagination.current - 1)"
            :disabled="pagination.current === 1"
            class="btn-secondary disabled:opacity-50"
          >
            前へ
          </button>
          <button
            @click="changePage(pagination.current + 1)"
            :disabled="pagination.current === pagination.pages"
            class="btn-secondary disabled:opacity-50"
          >
            次へ
          </button>
        </div>
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
          <div>
            <p class="text-sm text-gray-700">
              {{ pagination.total }}件中 {{ pagination.from }}〜{{ pagination.to }}件を表示
            </p>
          </div>
          <div>
            <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
              <button
                @click="changePage(pagination.current - 1)"
                :disabled="pagination.current === 1"
                class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
              >
                前へ
              </button>
              <button
                v-for="page in visiblePages"
                :key="page"
                @click="changePage(page)"
                :class="[
                  page === pagination.current
                    ? 'bg-blue-50 border-blue-500 text-blue-600'
                    : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50',
                  'relative inline-flex items-center px-4 py-2 border text-sm font-medium'
                ]"
              >
                {{ page }}
              </button>
              <button
                @click="changePage(pagination.current + 1)"
                :disabled="pagination.current === pagination.pages"
                class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50"
              >
                次へ
              </button>
            </nav>
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

const questions = ref([])
const categories = ref([])
const loading = ref(false)

const filters = reactive({
  search: '',
  category: '',
  difficulty: ''
})

const pagination = reactive({
  current: 1,
  pages: 1,
  total: 0,
  from: 0,
  to: 0,
  perPage: 20
})

const fetchQuestions = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams({
      page: pagination.current,
      per_page: pagination.perPage,
      ...filters
    })
    
    const response = await $fetch(`/api/questions?${params}`, {
      baseURL: config.public.apiBase,
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    
    questions.value = response.data || []
    Object.assign(pagination, response.pagination || {})
  } catch (error) {
    console.error('Questions fetch failed:', error)
  } finally {
    loading.value = false
  }
}

const fetchCategories = async () => {
  try {
    const response = await $fetch('/api/questions/categories', {
      baseURL: config.public.apiBase,
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      }
    })
    categories.value = response
  } catch (error) {
    console.error('Categories fetch failed:', error)
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
    await fetchQuestions()
  } catch (error) {
    console.error('Delete failed:', error)
    alert('削除に失敗しました。')
  }
}

const changePage = (page) => {
  if (page >= 1 && page <= pagination.pages) {
    pagination.current = page
    fetchQuestions()
  }
}

const resetFilters = () => {
  Object.assign(filters, {
    search: '',
    category: '',
    difficulty: ''
  })
  pagination.current = 1
  fetchQuestions()
}

const debouncedSearch = debounce(() => {
  pagination.current = 1
  fetchQuestions()
}, 500)

const visiblePages = computed(() => {
  const pages = []
  const start = Math.max(1, pagination.current - 2)
  const end = Math.min(pagination.pages, pagination.current + 2)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  
  return pages
})

const getDifficultyClass = (difficulty) => {
  switch (difficulty) {
    case 'easy': return 'bg-green-100 text-green-800'
    case 'medium': return 'bg-yellow-100 text-yellow-800'
    case 'hard': return 'bg-red-100 text-red-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

const getDifficultyLabel = (difficulty) => {
  switch (difficulty) {
    case 'easy': return '易しい'
    case 'medium': return '普通'
    case 'hard': return '難しい'
    default: return '不明'
  }
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString('ja-JP')
}

function debounce(func, wait) {
  let timeout
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout)
      func(...args)
    }
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
}

onMounted(() => {
  fetchQuestions()
  fetchCategories()
})
</script>