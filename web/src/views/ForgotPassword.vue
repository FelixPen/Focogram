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
  <div class="forgot-password-page">
    <div v-if="message.text" class="toast" :class="message.type">
      {{ message.text }}
    </div>
    
    <div class="forgot-card">
      <h2>找回密码</h2>
      <p class="subtitle">请输入注册时使用的邮箱</p>

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
        <div class="form-group">
          <label>注册邮箱</label>
          <input
            v-model="email"
            type="email"
            placeholder="请输入注册时的完整邮箱地址"
            @keyup.enter="verifyEmail"
          />
        </div>

        <button class="next-btn" :disabled="loading" @click="verifyEmail">
          下一步
        </button>
      </div>

      <div v-if="step === 2" class="step-content">
        <div class="email-ok">
          <span class="check-icon">✓</span>
          <p>邮箱验证成功！</p>
          <p class="email-text">{{ email }}</p>
        </div>

        <div class="form-group">
          <label>新密码</label>
          <input
            v-model="newPassword"
            type="password"
            placeholder="请设置新密码（至少6位）"
          />
        </div>

        <div class="form-group">
          <label>确认密码</label>
          <input
            v-model="confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            @keyup.enter="handleResetPassword"
          />
        </div>

        <div class="btn-group">
          <button class="back-btn" @click="step = 1">返回上一步</button>
          <button class="confirm-btn" :disabled="loading" @click="handleResetPassword">
            确认重置
          </button>
        </div>
      </div>

      <div class="back-to-login">
        <span @click="router.push('/login')">← 返回登录</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: 8px;
  color: white;
  font-weight: 600;
  z-index: 9999;
  animation: slideDown 0.3s ease;
}

.toast.success {
  background: #00ba7c;
}

.toast.warning {
  background: #f59e0b;
}

.toast.error {
  background: #f43f5e;
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

.forgot-password-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.forgot-card {
  background: white;
  border-radius: 16px;
  padding: 40px;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.forgot-card h2 {
  font-size: 28px;
  font-weight: 800;
  margin: 0 0 8px 0;
  text-align: center;
  color: #0f1419;
}

.subtitle {
  color: #536471;
  text-align: center;
  margin: 0 0 32px 0;
}

.steps-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 32px;
  gap: 12px;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.step-number {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #eff3f4;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  color: #536471;
  font-size: 14px;
  transition: all 0.3s;
}

.step.active .step-number {
  background: #1d9bf0;
  color: white;
}

.step.done .step-number {
  background: #00ba7c;
  color: white;
}

.step-text {
  font-size: 12px;
  color: #536471;
}

.step.active .step-text {
  color: #1d9bf0;
  font-weight: 600;
}

.step-line {
  width: 60px;
  height: 2px;
  background: #eff3f4;
  margin-bottom: 16px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #0f1419;
}

.form-group input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #eff3f4;
  border-radius: 8px;
  font-size: 15px;
  outline: none;
  transition: all 0.2s;
  box-sizing: border-box;
}

.form-group input:focus {
  border-color: #1d9bf0;
}

.email-ok {
  background: #ecfdf5;
  border: 1px solid #d1fae5;
  border-radius: 8px;
  padding: 16px;
  text-align: center;
  margin-bottom: 20px;
}

.check-icon {
  display: inline-block;
  width: 28px;
  height: 28px;
  background: #00ba7c;
  color: white;
  border-radius: 50%;
  font-weight: 700;
  line-height: 28px;
  margin-bottom: 8px;
}

.email-ok p {
  margin: 0;
  color: #065f46;
  font-size: 14px;
  font-weight: 600;
}

.email-text {
  font-size: 13px !important;
  font-weight: 500 !important;
  opacity: 0.8;
  margin-top: 4px !important;
}

.next-btn, .confirm-btn {
  width: 100%;
  padding: 14px;
  background: #1d9bf0;
  color: white;
  border: none;
  border-radius: 25px;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  margin-top: 8px;
}

.next-btn:hover, .confirm-btn:hover {
  background: #1a8cd8;
}

.next-btn:disabled, .confirm-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-group {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.back-btn {
  flex: 1;
  padding: 14px;
  background: #eff3f4;
  color: #0f1419;
  border: none;
  border-radius: 25px;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.back-btn:hover {
  background: #e2e6e9;
}

.confirm-btn {
  flex: 1;
  margin-top: 0;
}

.back-to-login {
  text-align: center;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #eff3f4;
}

.back-to-login span {
  color: #1d9bf0;
  cursor: pointer;
  font-weight: 600;
  font-size: 14px;
}

.back-to-login span:hover {
  text-decoration: underline;
}
</style>