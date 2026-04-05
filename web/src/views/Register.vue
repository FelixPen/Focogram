<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { register as handleRegister } from '../api/auth'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const form = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})
const error = ref('')
const success = ref('')
const loading = ref(false)

const handleSubmit = async () => {
  if (!form.value.username || !form.value.email || !form.value.password) {
    error.value = '请填写完整信息'
    return
  }
  if (form.value.password !== form.value.confirmPassword) {
    error.value = '两次密码输入不一致'
    return
  }

  error.value = ''
  loading.value = true
  try {
    const res = await handleRegister(form.value)
    
    // 注册成功自动登录 - 先存 token
    localStorage.setItem('token', res.token)
    localStorage.setItem('userid', res.user_id)
    
    // 拉取用户信息完成登录
    const userRes = await userStore.fetchUserInfo(res.user_id)
    userStore.setToken(res.token)
    userStore.setUser(userRes.user)
    
    success.value = `注册成功！您的账号是：${res.user_id}，已自动登录，正在进入首页...`
    setTimeout(() => {
      router.push('/')
    }, 1500)
  } catch (err) {
    error.value = err.response?.data?.error || '注册失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-container">
    <h1>Focogram</h1>
    <p style="margin-bottom: 24px; color: #8e8e8e;">注册以开始你的社交之旅</p>

    <div v-if="error" class="error-message">{{ error }}</div>
    <div v-if="success" class="success-message">{{ success }}</div>

    <form @submit.prevent="handleSubmit">
      <div class="form-group">
        <label>用户名</label>
        <input
          v-model="form.username"
          type="text"
          placeholder="请输入用户名"
          autocomplete="username"
        />
      </div>

      <div class="form-group">
        <label>邮箱</label>
        <input
          v-model="form.email"
          type="email"
          placeholder="请输入邮箱"
          autocomplete="email"
        />
      </div>

      <div class="form-group">
        <label>密码</label>
        <input
          v-model="form.password"
          type="password"
          placeholder="请输入密码"
          autocomplete="new-password"
        />
      </div>

      <div class="form-group">
        <label>确认密码</label>
        <input
          v-model="form.confirmPassword"
          type="password"
          placeholder="请再次输入密码"
          autocomplete="new-password"
        />
      </div>

      <button type="submit" class="btn btn-primary" :disabled="loading">
        {{ loading ? '注册中...' : '注册' }}
      </button>
    </form>

    <div class="auth-link">
      已有账号？ <router-link to="/login">立即登录</router-link>
    </div>
  </div>
</template>
