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
const showPassword = ref(false)

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
  <div class="auth-page">
    <div class="auth-background">
      <div class="gradient-orb orb-1"></div>
      <div class="gradient-orb orb-2"></div>
      <div class="gradient-orb orb-3"></div>
    </div>

    <div class="auth-card">
      <div class="card-header">
        <div class="logo-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <circle cx="12" cy="12" r="3"></circle>
            <line x1="12" y1="2" x2="12" y2="4"></line>
            <line x1="12" y1="20" x2="12" y2="22"></line>
            <line x1="2" y1="12" x2="4" y2="12"></line>
            <line x1="20" y1="12" x2="22" y2="12"></line>
          </svg>
        </div>
        <h1 class="app-title">Focogram</h1>
        <p class="app-subtitle">专注你的社交世界</p>
      </div>

      <div v-if="error" class="error-toast">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        {{ error }}
      </div>

      <form @submit.prevent="handleSubmit" class="auth-form">
        <div class="input-group">
          <div class="input-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
          </div>
          <input
            v-model="form.userid"
            type="text"
            placeholder="账号"
            autocomplete="username"
            class="auth-input"
          />
        </div>

        <div class="input-group">
          <div class="input-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
            </svg>
          </div>
          <input
            v-model="form.password"
            :type="showPassword ? 'text' : 'password'"
            placeholder="密码"
            autocomplete="current-password"
            class="auth-input"
          />
          <button type="button" class="toggle-password" @click="showPassword = !showPassword">
            <svg v-if="showPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
              <circle cx="12" cy="12" r="3"></circle>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
              <line x1="1" y1="1" x2="23" y2="23"></line>
            </svg>
          </button>
        </div>

        <button type="submit" class="submit-btn" :disabled="loading">
          <span v-if="!loading">登录</span>
          <span v-else class="loading-spinner">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" stroke-dasharray="60" stroke-dashoffset="20"></circle>
            </svg>
            登录中...
          </span>
        </button>
      </form>

      <div class="forgot-link">
        <router-link to="/forgot-password">忘记密码？</router-link>
      </div>

      <div class="divider">
        <span>或</span>
      </div>

      <div class="auth-footer">
        <p>还没有账号？</p>
        <router-link to="/register" class="register-link">立即注册</router-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #f09433 0%, #e6683c 25%, #dc2743 50%, #cc2366 75%, #bc1888 100%);
}

.auth-background {
  position: absolute;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.gradient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.6;
  animation: float 20s ease-in-out infinite;
}

.orb-1 {
  width: 500px;
  height: 500px;
  background: linear-gradient(135deg, #f09433, #e6683c);
  top: -200px;
  left: -100px;
  animation-delay: 0s;
}

.orb-2 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, #dc2743, #cc2366);
  bottom: -150px;
  right: -100px;
  animation-delay: -7s;
}

.orb-3 {
  width: 350px;
  height: 350px;
  background: linear-gradient(135deg, #bc1888, #e1306c);
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation-delay: -14s;
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  25% {
    transform: translate(50px, -50px) scale(1.1);
  }
  50% {
    transform: translate(-30px, 30px) scale(0.95);
  }
  75% {
    transform: translate(-50px, -30px) scale(1.05);
  }
}

.auth-card {
  position: relative;
  width: 100%;
  max-width: 420px;
  padding: 48px 40px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-radius: 32px;
  box-shadow: 
    0 20px 60px rgba(0, 0, 0, 0.15),
    0 0 0 1px rgba(255, 255, 255, 0.5) inset;
  animation: cardEntry 0.8s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  margin: 20px;
}

@keyframes cardEntry {
  0% {
    opacity: 0;
    transform: scale(0.8) translateY(30px);
  }
  70% {
    transform: scale(1.02) translateY(-5px);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.card-header {
  text-align: center;
  margin-bottom: 32px;
  animation: fadeInUp 0.6s ease 0.2s both;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.logo-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 16px;
  background: linear-gradient(135deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 10px 30px rgba(225, 48, 108, 0.3);
  animation: logoFloat 3s ease-in-out infinite;
}

@keyframes logoFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-8px);
  }
}

.logo-icon svg {
  width: 32px;
  height: 32px;
}

.app-title {
  font-size: 32px;
  font-weight: 800;
  margin: 0 0 8px;
  background: linear-gradient(135deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.app-subtitle {
  font-size: 14px;
  color: #8e8e8e;
  margin: 0;
}

.error-toast {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #fee2e2, #fef2f2);
  border: 1px solid #fecaca;
  border-radius: 12px;
  color: #dc2626;
  font-size: 13px;
  margin-bottom: 20px;
  animation: shake 0.5s ease;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}

.error-toast svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.auth-form {
  animation: fadeInUp 0.6s ease 0.3s both;
}

.input-group {
  position: relative;
  margin-bottom: 16px;
}

.input-icon {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  color: #9ca3af;
  pointer-events: none;
  transition: color 0.3s;
}

.input-icon svg {
  width: 100%;
  height: 100%;
}

.auth-input {
  width: 100%;
  padding: 16px 48px;
  border: 2px solid #e5e7eb;
  border-radius: 16px;
  font-size: 15px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: #f9fafb;
  box-sizing: border-box;
}

.auth-input:focus {
  outline: none;
  border-color: #e1306c;
  background: white;
  box-shadow: 0 0 0 4px rgba(225, 48, 108, 0.1);
}

.auth-input:focus + .input-icon,
.auth-input:focus ~ .input-icon {
  color: #e1306c;
}

.input-group:has(.auth-input:focus) .input-icon {
  color: #e1306c;
}

.toggle-password {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  background: none;
  border: none;
  cursor: pointer;
  color: #9ca3af;
  transition: color 0.3s;
  padding: 0;
}

.toggle-password:hover {
  color: #e1306c;
}

.toggle-password svg {
  width: 100%;
  height: 100%;
}

.submit-btn {
  width: 100%;
  padding: 16px;
  background: linear-gradient(135deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888);
  color: white;
  border: none;
  border-radius: 16px;
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 10px 30px rgba(225, 48, 108, 0.3);
  margin-top: 8px;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 15px 40px rgba(225, 48, 108, 0.4);
}

.submit-btn:active:not(:disabled) {
  transform: translateY(0);
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.loading-spinner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.loading-spinner svg {
  width: 18px;
  height: 18px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.forgot-link {
  text-align: right;
  margin-top: 16px;
  animation: fadeInUp 0.6s ease 0.4s both;
}

.forgot-link a {
  color: #e1306c;
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  transition: color 0.3s;
}

.forgot-link a:hover {
  color: #c13584;
  text-decoration: underline;
}

.divider {
  display: flex;
  align-items: center;
  margin: 24px 0;
  animation: fadeInUp 0.6s ease 0.5s both;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: linear-gradient(to right, transparent, #e5e7eb, transparent);
}

.divider span {
  padding: 0 16px;
  color: #9ca3af;
  font-size: 13px;
  font-weight: 500;
}

.auth-footer {
  text-align: center;
  animation: fadeInUp 0.6s ease 0.6s both;
}

.auth-footer p {
  margin: 0 0 8px;
  color: #6b7280;
  font-size: 14px;
}

.register-link {
  display: inline-block;
  padding: 12px 32px;
  background: transparent;
  color: #e1306c;
  border: 2px solid #e1306c;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.register-link:hover {
  background: linear-gradient(135deg, #f09433, #e6683c, #dc2743);
  color: white;
  border-color: transparent;
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(225, 48, 108, 0.3);
}

:deep(.dark) .auth-page {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f0f1e 100%);
}

:deep(.dark) .auth-card {
  background: rgba(30, 30, 40, 0.95);
  box-shadow: 
    0 20px 60px rgba(0, 0, 0, 0.5),
    0 0 0 1px rgba(255, 255, 255, 0.1) inset;
}

:deep(.dark) .app-subtitle {
  color: #9ca3af;
}

:deep(.dark) .auth-input {
  background: #1f2937;
  border-color: #374151;
  color: #f9fafb;
}

:deep(.dark) .auth-input:focus {
  background: #1f2937;
  border-color: #e1306c;
  box-shadow: 0 0 0 4px rgba(225, 48, 108, 0.1);
}

:deep(.dark) .divider::before,
:deep(.dark) .divider::after {
  background: linear-gradient(to right, transparent, #374151, transparent);
}

:deep(.dark) .auth-footer p {
  color: #9ca3af;
}

:deep(.dark) .register-link {
  color: #e1306c;
  border-color: #e1306c;
}

:deep(.dark) .register-link:hover {
  background: linear-gradient(135deg, #f09433, #e6683c, #dc2743);
  color: white;
  border-color: transparent;
}
</style>
