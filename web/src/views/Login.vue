<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const form = ref({
  userid: '',
  password: ''
})
const error = ref('')
const loading = ref(false)

const handleSubmit = async () => {
  if (!form.value.userid || !form.value.password) {
    error.value = '请填写所有字段'
    return
  }

  loading.value = true
  error.value = ''

  try {
    await userStore.handleLogin(form.value)
    router.push('/')
  } catch (err) {
    error.value = err.response?.data?.error || '登录失败，请检查账号和密码'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <h1>Focogram</h1>
    <p style="margin-bottom: 24px; color: #8e8e8e;">专注你的社交世界</p>

    <div v-if="error" class="error-message">{{ error }}</div>

    <form @submit.prevent="handleSubmit">
      <div class="form-group">
        <label>账号</label>
        <input
          v-model="form.userid"
          type="text"
          placeholder="请输入账号"
          autocomplete="username"
        />
      </div>

      <div class="form-group">
        <label>密码</label>
        <input
          v-model="form.password"
          type="password"
          placeholder="请输入密码"
          autocomplete="current-password"
        />
      </div>

      <button type="submit" class="btn btn-primary" :disabled="loading">
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </form>

    <div class="forgot-password">
      <router-link to="/forgot-password">忘记密码？</router-link>
    </div>

    <div class="auth-link">
      还没有账号？ <router-link to="/register">立即注册</router-link>
    </div>
  </div>
</template>
