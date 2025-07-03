<template>
  <div class="max-w-4xl mx-auto">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-gray-900">新しい問題を作成</h1>
      <NuxtLink to="/questions" class="btn-secondary">
        問題一覧に戻る
      </NuxtLink>
    </div>

    <form @submit.prevent="handleSubmit" class="space-y-6">
      <div class="card">
        <h2 class="text-xl font-semibold text-gray-900 mb-4">基本情報</h2>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label for="category" class="block text-sm font-medium text-gray-700 mb-2">
              カテゴリ
            </label>
            <input
              id="category"
              v-model="form.category"
              type="text"
              required
              class="form-input"
              placeholder="例: 一般知識"
            />
          </div>
          
          <div>
            <label for="difficulty" class="block text-sm font-medium text-gray-700 mb-2">
              難易度
            </label>
            <select id="difficulty" v-model="form.difficulty" required class="form-input">
              <option value="">選択してください</option>
              <option value="easy">易しい</option>
              <option value="medium">普通</option>
              <option value="hard">難しい</option>
            </select>
          </div>
        </div>

        <div class="mt-6">
          <label for="question" class="block text-sm font-medium text-gray-700 mb-2">
            問題文 *
          </label>
          <textarea
            id="question"
            v-model="form.question_text"
            required
            rows="4"
            class="form-textarea"
            placeholder="問題文を入力してください"
          ></textarea>
        </div>

        <div class="mt-6">
          <label class="block text-sm font-medium text-gray-700 mb-2">
            画像（オプション）
          </label>
          <div class="flex items-center space-x-4">
            <input
              ref="fileInput"
              type="file"
              accept="image/*"
              @change="handleFileChange"
              class="hidden"
            />
            <button
              type="button"
              @click="$refs.fileInput.click()"
              class="btn-secondary"
            >
              画像を選択
            </button>
            <span v-if="form.image_file" class="text-sm text-gray-600">
              {{ form.image_file.name }}
            </span>
          </div>
          
          <div v-if="imagePreview" class="mt-4">
            <img :src="imagePreview" alt="プレビュー" class="max-w-xs rounded-lg shadow-sm" />
          </div>
        </div>
      </div>

      <div class="card">
        <h2 class="text-xl font-semibold text-gray-900 mb-4">選択肢</h2>
        
        <div class="space-y-4">
          <div v-for="(option, index) in form.options" :key="index" class="flex items-center space-x-4">
            <span class="flex-shrink-0 w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center text-sm font-medium">
              {{ String.fromCharCode(65 + index) }}
            </span>
            <input
              v-model="option.text"
              type="text"
              required
              class="form-input flex-1"
              :placeholder="`選択肢 ${String.fromCharCode(65 + index)}`"
            />
            <label class="flex items-center">
              <input
                v-model="form.correct_answer"
                type="radio"
                :value="index"
                class="mr-2"
                required
              />
              <span class="text-sm text-gray-600">正解</span>
            </label>
          </div>
        </div>
      </div>

      <div class="card">
        <h2 class="text-xl font-semibold text-gray-900 mb-4">詳細設定</h2>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label for="time_limit" class="block text-sm font-medium text-gray-700 mb-2">
              制限時間（秒）
            </label>
            <input
              id="time_limit"
              v-model.number="form.time_limit"
              type="number"
              min="10"
              max="300"
              class="form-input"
              placeholder="60"
            />
          </div>
          
          <div>
            <label for="points" class="block text-sm font-medium text-gray-700 mb-2">
              配点
            </label>
            <input
              id="points"
              v-model.number="form.points"
              type="number"
              min="1"
              max="100"
              class="form-input"
              placeholder="10"
            />
          </div>
        </div>

        <div class="mt-6">
          <label for="explanation" class="block text-sm font-medium text-gray-700 mb-2">
            解説（オプション）
          </label>
          <textarea
            id="explanation"
            v-model="form.explanation"
            rows="3"
            class="form-textarea"
            placeholder="解説を入力してください"
          ></textarea>
        </div>
      </div>

      <div class="flex justify-end space-x-4">
        <NuxtLink to="/questions" class="btn-secondary">
          キャンセル
        </NuxtLink>
        <button type="submit" :disabled="loading" class="btn-primary">
          {{ loading ? '作成中...' : '問題を作成' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
definePageMeta({
  middleware: 'auth'
})

const router = useRouter()
const { $fetch } = useNuxtApp()
const config = useRuntimeConfig()

const form = reactive({
  category: '',
  difficulty: '',
  question_text: '',
  options: [
    { text: '' },
    { text: '' },
    { text: '' },
    { text: '' }
  ],
  correct_answer: null,
  time_limit: 60,
  points: 10,
  explanation: '',
  image_file: null
})

const loading = ref(false)
const imagePreview = ref(null)

const handleFileChange = (event) => {
  const file = event.target.files[0]
  if (file) {
    form.image_file = file
    const reader = new FileReader()
    reader.onload = (e) => {
      imagePreview.value = e.target.result
    }
    reader.readAsDataURL(file)
  }
}

const handleSubmit = async () => {
  loading.value = true
  
  try {
    const formData = new FormData()
    
    const questionData = {
      category: form.category,
      difficulty: form.difficulty,
      question_text: form.question_text,
      options: form.options.map(opt => opt.text),
      correct_answer: parseInt(form.correct_answer),
      time_limit: form.time_limit,
      points: form.points,
      explanation: form.explanation
    }
    
    formData.append('question', JSON.stringify(questionData))
    
    if (form.image_file) {
      formData.append('image', form.image_file)
    }
    
    await $fetch('/api/questions', {
      baseURL: config.public.apiBase,
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('admin_token')}`
      },
      body: formData
    })
    
    router.push('/questions')
  } catch (error) {
    console.error('Question creation failed:', error)
    alert('問題の作成に失敗しました。')
  } finally {
    loading.value = false
  }
}
</script>