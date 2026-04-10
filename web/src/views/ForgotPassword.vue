<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { resetPasswordByEmail } from '../api/auth'

const router = useRouter()

const step = ref(1)
const email = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const message = ref({ type: '', text: '' })
const showPassword = ref(false)
const showConfirmPassword = ref(false)

const showMessage = (type, text) => {
  message.value = { type, text }
  setTimeout(() => {
    message.value = { type: '', text: '' }
  }, 3000)
}

const verifyEmail = () => {
  if (!email.value) {
    showMessage('warning', '请输入邮箱地址')
    return
  }
  
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(email.value)) {
    showMessage('warning', '请输入有效的邮箱地址')
    return
  }
  
  step.value = 2
}

const handleResetPassword = async () => {
  if (!newPassword.value || newPassword.value.length < 6) {
    showMessage('warning', '新密码至少6位')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    showMessage('warning', '两次密码输入不一致')
    return
  }

  loading.value = true
  try {
    await resetPasswordByEmail({
      email: email.value,
      new_password: newPassword.value
    })
    showMessage('success', '密码重置成功！')
    setTimeout(() => {
      router.push('/login')
    }, 1500)
  } catch (err) {
    showMessage('error', err.response?.data?.error || '重置失败')
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

    <div v-if="message.text" class="toast" :class="message.type">
      <svg v-if="message.type === 'success'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
        <polyline points="22 4 12 14.01 9 11.01"></polyline>
      </svg>
      <svg v-else-if="message.type === 'warning'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
        <line x1="12" y1="9" x2="12" y2="13"></line>
        <line x1="12" y1="17" x2="12.01" y2="17"></line>
      </svg>
      <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <line x1="12" y1="8" x2="12" y2="12"></line>
        <line x1="12" y1="16" x2="12.01" y2="16"></line>
      </svg>
      {{ message.text }}
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
        <h1 class="app-title">找回密码</h1>
        <p class="app-subtitle">请输入注册时使用的邮箱</p>
      </div>

      <div class="steps-indicator">
        <div class="step" :class="{ active: step === 1, done: step > 1 }">
          <span class="step-number">1</span>
          <span class="step-text">验证邮箱</span>
        </div>
        <div class="step-line"></div>
        <div class="step" :class="{ active: step === 2 }">
          <span class="step-number">2</span>
          <span class="step-text">设置密码</span>
        </div>
      </div>

      <div v-if="step === 1" class="step-content">
        <div class="input-group">
          <div class="input-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
              <polyline points="22,6 12,13 2,6"></polyline>
            </svg>
          </div>
          <input
            v-model="email"
            type="email"
            placeholder="注册邮箱"
            class="auth-input"
            @keyup.enter="verifyEmail"
          />
        </div>

        <button class="submit-btn" :disabled="loading" @click="verifyEmail">
          下一步
        </button>
      </div>

      <div v-if="step === 2" class="step-content">
        <div class="email-ok">
          <div class="check-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
          </div>
          <p class="email-title">邮箱验证成功！</p>
          <p class="email-text">{{ email }}</p>
        </div>

        <div class="input-group">
          <div class="input-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
            </svg>
          </div>
          <input
            v-model="newPassword"
            :type="showPassword ? 'text' : 'password'"
            placeholder="新密码（至少6位）"
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

        <div class="input-group">
          <div class="input-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
            </svg>
          </div>
          <input
            v-model="confirmPassword"
            :type="showConfirmPassword ? 'text' : 'password'"
            placeholder="确认新密码"
            class="auth-input"
            @keyup.enter="handleResetPassword"
          />
          <button type="button" class="toggle-password" @click="showConfirmPassword = !showConfirmPassword">
            <svg v-if="showConfirmPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
              <circle cx="12" cy="12" r="3"></circle>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
              <line x1="1" y1="1" x2="23" y2="23"></line>
            </svg>
          </button>
        </div>

        <div class="btn-group">
          <button class="back-btn" @click="step = 1">返回</button>
          <button class="submit-btn" :disabled="loading" @click="handleResetPassword">
            <span v-if="!loading">确认重置</span>
            <span v-else class="loading-spinner">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10" stroke-dasharray="60" stroke-dashoffset="20"></circle>
              </svg>
              重置中...
            </span>
          </button>
        </div>
      </div>

      <div class="back-to-login">
        <router-link to="/login">← 返回登录</router-link>
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

.toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border-radius: 12px;
  color: white;
  font-weight: 600;
  font-size: 14px;
  z-index: 9999;
  animation: slideDown 0.3s ease;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
}

.toast svg {
  width: 18px;
  height: 18px;
}

.toast.success {
  background: linear-gradient(135deg, #10b981, #059669);
}

.toast.warning {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

.toast.error {
  background: linear-gradient(135deg, #ef4444, #dc2626);
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
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

.steps-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 32px;
  gap: 12px;
  animation: fadeInUp 0.6s ease 0.3s both;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.step-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  color: #9ca3af;
  font-size: 15px;
  transition: all 0.3s;
}

.step.active .step-number {
  background: linear-gradient(135deg, #f09433, #e6683c, #dc2743);
  color: white;
  box-shadow: 0 5px 15px rgba(225, 48, 108, 0.3);
}

.step.done .step-number {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
}

.step-text {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
}

.step.active .step-text {
  color: #e1306c;
  font-weight: 600;
}

.step-line {
  width: 60px;
  height: 2px;
  background: linear-gradient(to right, #f3f4f6, #e5e7eb, #f3f4f6);
  margin-bottom: 20px;
}

.step-content {
  animation: fadeInUp 0.6s ease 0.4s both;
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

.email-ok {
  background: linear-gradient(135deg, #d1fae5, #ecfdf5);
  border: 1px solid #a7f3d0;
  border-radius: 16px;
  padding: 20px;
  text-align: center;
  margin-bottom: 20px;
}

.check-icon {
  width: 48px;
  height: 48px;
  margin: 0 auto 12px;
  background: linear-gradient(135deg, #10b981, #059669);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 5px 15px rgba(16, 185, 129, 0.3);
}

.check-icon svg {
  width: 24px;
  height: 24px;
}

.email-title {
  margin: 0 0 4px;
  color: #065f46;
  font-size: 15px;
  font-weight: 600;
}

.email-text {
  margin: 0;
  color: #059669;
  font-size: 13px;
  font-weight: 500;
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

.btn-group {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.back-btn {
  flex: 1;
  padding: 16px;
  background: #f3f4f6;
  color: #6b7280;
  border: none;
  border-radius: 16px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.back-btn:hover {
  background: #e5e7eb;
}

.back-to-login {
  text-align: center;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #e5e7eb;
  animation: fadeInUp 0.6s ease 0.5s both;
}

.back-to-login a {
  color: #e1306c;
  font-weight: 600;
  font-size: 14px;
  text-decoration: none;
  transition: color 0.3s;
}

.back-to-login a:hover {
  color: #c13584;
  text-decoration: underline;
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

:deep(.dark) .step-number {
  background: #374151;
  color: #9ca3af;
}

:deep(.dark) .step-line {
  background: linear-gradient(to right, #374151, #4b5563, #374151);
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

:deep(.dark) .email-ok {
  background: rgba(16, 185, 129, 0.1);
  border-color: rgba(16, 185, 129, 0.3);
}

:deep(.dark) .email-title {
  color: #34d399;
}

:deep(.dark) .email-text {
  color: #6ee7b7;
}

:deep(.dark) .back-btn {
  background: #374151;
  color: #d1d5db;
}

:deep(.dark) .back-btn:hover {
  background: #4b5563;
}

:deep(.dark) .back-to-login {
  border-top-color: #374151;
}

:deep(.dark) .back-to-login a {
  color: #e1306c;
}

:deep(.dark) .back-to-login a:hover {
  color: #f472b6;
}
</style>
