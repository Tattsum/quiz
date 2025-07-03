<template>
  <div class="min-h-screen bg-gray-50">
    <nav class="bg-white shadow-sm border-b" v-if="isLoggedIn">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center">
            <h1 class="text-lg sm:text-xl font-semibold text-gray-900">クイズ管理システム</h1>
            
            <!-- モバイルメニューボタン -->
            <button 
              @click="mobileMenuOpen = !mobileMenuOpen"
              class="ml-auto md:hidden inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path v-if="!mobileMenuOpen" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          
          <!-- デスクトップメニュー -->
          <div class="hidden md:flex items-center space-x-4">
            <NuxtLink to="/dashboard" class="text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium">
              ダッシュボード
            </NuxtLink>
            <NuxtLink to="/questions" class="text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium">
              問題管理
            </NuxtLink>
            <NuxtLink to="/quiz-control" class="text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium">
              クイズ制御
            </NuxtLink>
            <NuxtLink to="/ranking" class="text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-sm font-medium">
              ランキング
            </NuxtLink>
            <button @click="logout" class="btn-secondary">
              ログアウト
            </button>
          </div>
        </div>
        
        <!-- モバイルメニュー -->
        <div v-if="mobileMenuOpen" class="md:hidden">
          <div class="px-2 pt-2 pb-3 space-y-1 sm:px-3 border-t border-gray-200">
            <NuxtLink 
              to="/dashboard" 
              @click="mobileMenuOpen = false"
              class="block text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-base font-medium"
            >
              ダッシュボード
            </NuxtLink>
            <NuxtLink 
              to="/questions" 
              @click="mobileMenuOpen = false"
              class="block text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-base font-medium"
            >
              問題管理
            </NuxtLink>
            <NuxtLink 
              to="/quiz-control" 
              @click="mobileMenuOpen = false"
              class="block text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-base font-medium"
            >
              クイズ制御
            </NuxtLink>
            <NuxtLink 
              to="/ranking" 
              @click="mobileMenuOpen = false"
              class="block text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-base font-medium"
            >
              ランキング
            </NuxtLink>
            <button 
              @click="logout" 
              class="block w-full text-left text-gray-700 hover:text-blue-600 px-3 py-2 rounded-md text-base font-medium"
            >
              ログアウト
            </button>
          </div>
        </div>
      </div>
    </nav>
    
    <main class="max-w-7xl mx-auto py-4 sm:py-6 px-4 sm:px-6 lg:px-8">
      <slot />
    </main>
  </div>
</template>

<script setup>
const router = useRouter()
const route = useRoute()

const mobileMenuOpen = ref(false)

const isLoggedIn = computed(() => {
  return route.path !== '/login'
})

const logout = () => {
  // ログアウト処理
  localStorage.removeItem('admin_token')
  mobileMenuOpen.value = false
  router.push('/login')
}

// ルート変更時にモバイルメニューを閉じる
watch(route, () => {
  mobileMenuOpen.value = false
})
</script>