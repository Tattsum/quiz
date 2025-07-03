<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          管理者ログイン
        </h2>
        <p class="mt-2 text-center text-sm text-gray-600">
          クイズ管理システムにログインしてください
        </p>
      </div>
      
      <form class="mt-8 space-y-6" @submit.prevent="handleLogin">
        <div class="rounded-md shadow-sm -space-y-px">
          <div>
            <label for="username" class="sr-only">ユーザー名</label>
            <input
              id="username"
              v-model="loginForm.username"
              name="username"
              type="text"
              required
              class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
              placeholder="ユーザー名"
            />
          </div>
          <div>
            <label for="password" class="sr-only">パスワード</label>
            <input
              id="password"
              v-model="loginForm.password"
              name="password"
              type="password"
              required
              class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
              placeholder="パスワード"
            />
          </div>
        </div>

        <div v-if="error" class="text-red-600 text-sm text-center">
          {{ error }}
        </div>

        <div>
          <button
            type="submit"
            :disabled="loading"
            class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span v-if="loading">ログイン中...</span>
            <span v-else>ログイン</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
definePageMeta({
  layout: false
})

const router = useRouter()
const { $fetch } = useNuxtApp()
const config = useRuntimeConfig()

const loginForm = reactive({
  username: '',
  password: ''
})

const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const response = await $fetch('/api/admin/login', {
      baseURL: config.public.apiBase,
      method: 'POST',
      body: {
        username: loginForm.username,
        password: loginForm.password
      }
    })
    
    if (response.token) {
      localStorage.setItem('admin_token', response.token)
      router.push('/dashboard')
    }
  } catch (err) {
    error.value = 'ログインに失敗しました。ユーザー名とパスワードを確認してください。'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    router.push('/dashboard')
  }
})
</script>