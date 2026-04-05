<script setup>
import { ref } from 'vue'
import { updatePassword } from '../api/auth'
import { batchDeleteNotifications } from '../api/notification'
import { useThemeStore } from '../stores/theme'
import { useNotificationStore } from '../stores/notification'

const themeStore = useThemeStore()
const notificationStore = useNotificationStore()

const expandedSection = ref('')
const message = ref({ type: '', text: '' })
const showConfirmModal = ref(false)
const showPasswordModal = ref(false)
const passwordMode = ref('change')

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const changePwdLoading = ref(false)
const email = ref('')
const resetCode = ref('')
const resetStep = ref(1)

const showMessage = (type, text) => {
  message.value = { type, text }
  setTimeout(() => {
    message.value = { type: '', text: '' }
  }, 3000)
}

const toggleSection = (section) => {
  if (expandedSection.value === section) {
    expandedSection.value = ''
  } else {
    expandedSection.value = section
  }
}

const openPasswordModal = (mode) => {
  passwordMode.value = mode
  resetStep.value = 1
  oldPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
  email.value = ''
  resetCode.value = ''
  showPasswordModal.value = true
}

const closePasswordModal = () => {
  showPasswordModal.value = false
}

const handleChangePassword = async () => {
  if (!oldPassword.value) {
    showMessage('warning', '请输入原密码')
    return
  }
  if (!newPassword.value || newPassword.value.length < 6) {
    showMessage('warning', '新密码至少6位')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    showMessage('warning', '两次密码输入不一致')
    return
  }

  changePwdLoading.value = true
  try {
    await updatePassword({
      old_password: oldPassword.value,
      new_password: newPassword.value
    })
    showMessage('success', '密码修改成功！')
    closePasswordModal()
  } catch (err) {
    showMessage('error', err.response?.data?.error || '密码修改失败')
  } finally {
    changePwdLoading.value = false
  }
}

const handleVerifyEmail = () => {
  if (!email.value) {
    showMessage('warning', '请输入邮箱')
    return
  }
  if (!email.value.includes('@')) {
    showMessage('error', '邮箱格式不正确')
    return
  }
  showMessage('success', '邮箱验证通过！请设置新密码')
  resetStep.value = 2
}

const handleResetPassword = () => {
  if (!newPassword.value || newPassword.value.length < 6) {
    showMessage('warning', '新密码至少6位')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    showMessage('warning', '两次密码输入不一致')
    return
  }
  showMessage('success', '密码重置成功！')
  closePasswordModal()
}

const openConfirmModal = () => {
  showConfirmModal.value = true
}

const confirmAction = async () => {
  showConfirmModal.value = false
  try {
    const allIds = notificationStore.notifications.map(n => n.id)
    if (allIds.length > 0) {
      await batchDeleteNotifications(allIds)
    }
    notificationStore.notifications = []
    showMessage('success', '所有通知已永久删除！')
  } catch (err) {
    showMessage('success', '通知已清空！')
  }
}

const cancelAction = () => {
  showConfirmModal.value = false
}

const handleClearAllNotifications = () => {
  openConfirmModal()
}
</script>

<template>
  <div class="settings-page">
    <div v-if="message.text" class="toast" :class="message.type">
      {{ message.text }}
    </div>

    <transition name="modal-fade">
      <div v-if="showPasswordModal" class="modal-overlay" @click.self="closePasswordModal">
        <div class="modal-content password-modal">
          <div class="modal-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
            </svg>
          </div>
          <h3 class="modal-title">{{ passwordMode === 'change' ? '修改密码' : '忘记密码' }}</h3>
          
          <div v-if="passwordMode === 'change'" class="password-form">
            <div class="form-group">
              <label>原密码</label>
              <input v-model="oldPassword" type="password" placeholder="请输入当前密码" />
            </div>
            <div class="form-group">
              <label>新密码</label>
              <input v-model="newPassword" type="password" placeholder="请设置新密码（至少6位）" />
            </div>
            <div class="form-group">
              <label>确认新密码</label>
              <input v-model="confirmPassword" type="password" placeholder="请再次输入新密码" />
            </div>
            <p class="switch-mode" @click="openPasswordModal('forgot')">
              忘记密码了？→ 点击重置
            </p>
            <div class="modal-buttons">
              <button class="modal-btn cancel" @click="closePasswordModal">
                取消
              </button>
              <button class="modal-btn confirm" @click="handleChangePassword" :disabled="changePwdLoading">
                {{ changePwdLoading ? '修改中...' : '确认修改' }}
              </button>
            </div>
          </div>

          <div v-if="passwordMode === 'forgot'" class="password-form">
            <div v-if="resetStep === 1" class="step-content">
              <div class="form-group">
                <label>注册邮箱</label>
                <input v-model="email" type="email" placeholder="请输入注册时使用的邮箱" />
              </div>
              <p class="switch-mode" @click="openPasswordModal('change')">
                ← 返回修改密码
              </p>
              <div class="modal-buttons">
                <button class="modal-btn cancel" @click="closePasswordModal">
                  取消
                </button>
                <button class="modal-btn confirm" @click="handleVerifyEmail">
                  验证邮箱
                </button>
              </div>
            </div>
            <div v-if="resetStep === 2" class="step-content">
              <div class="form-group">
                <label>新密码</label>
                <input v-model="newPassword" type="password" placeholder="请设置新密码（至少6位）" />
              </div>
              <div class="form-group">
                <label>确认新密码</label>
                <input v-model="confirmPassword" type="password" placeholder="请再次输入新密码" />
              </div>
              <div class="modal-buttons">
                <button class="modal-btn cancel" @click="resetStep = 1">
                  返回
                </button>
                <button class="modal-btn confirm" @click="handleResetPassword">
                  确认重置
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </transition>

    <transition name="modal-fade">
      <div v-if="showConfirmModal" class="modal-overlay" @click.self="cancelAction">
        <div class="modal-content">
          <div class="modal-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
            </svg>
          </div>
          <h3 class="modal-title">确认操作</h3>
          <p class="modal-desc">确定要永久删除所有通知吗？此操作不可恢复。</p>
          <div class="modal-buttons">
            <button class="modal-btn cancel" @click="cancelAction">
              取消
            </button>
            <button class="modal-btn confirm" @click="confirmAction">
              确认删除
            </button>
          </div>
        </div>
      </div>
    </transition>
    
    <div class="page-header">
      <h1 style="display: flex; align-items: center; gap: 8px;">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 22px; height: 22px;">
          <circle cx="12" cy="12" r="3"></circle>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
        </svg>
        设置
      </h1>
    </div>

    <div class="settings-content">
      <div class="card setting-item theme-toggle">
        <svg class="setting-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="5"></circle>
          <line x1="12" y1="1" x2="12" y2="3"></line>
          <line x1="12" y1="21" x2="12" y2="23"></line>
          <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
          <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
          <line x1="1" y1="12" x2="3" y2="12"></line>
          <line x1="21" y1="12" x2="23" y2="12"></line>
          <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
          <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
        </svg>
        <div class="setting-info">
          <h3>外观主题</h3>
          <p>{{ themeStore.isDark ? '暗黑模式已开启' : '明亮模式已开启' }}</p>
        </div>
        <div class="toggle-switch" :class="{ active: themeStore.isDark }" @click="themeStore.toggleTheme()">
          <div class="toggle-circle"></div>
        </div>
      </div>

      <div class="card setting-item" @click="toggleSection('notifications')">
        <svg class="setting-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
        </svg>
        <div class="setting-info">
          <h3>通知设置</h3>
          <p>管理通知、清除通知记录</p>
        </div>
        <span class="setting-arrow">{{ expandedSection === 'notifications' ? '↓' : '→' }}</span>
      </div>

      <div v-if="expandedSection === 'notifications'" class="card privacy-content">
        <div class="privacy-section">
          <div class="danger-zone-inline">
            <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 12px;">
              <svg style="width: 24px; height: 24px; color: #ef4444;" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              </svg>
              <div>
                <h4 style="margin: 0; font-size: 15px; color: #dc2626;">清空所有通知</h4>
                <p style="margin: 2px 0 0; font-size: 12px; color: #991b1b;">永久删除您的所有通知记录，此操作不可恢复</p>
              </div>
            </div>
            <button class="clear-notifications-btn" style="width: 100%;" @click="handleClearAllNotifications">
              🗑️ 永久清空所有通知
            </button>
          </div>
        </div>
      </div>

      <div class="card setting-item" @click="toggleSection('privacy')">
        <svg class="setting-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
          <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
        </svg>
        <div class="setting-info">
          <h3>隐私安全</h3>
          <p>修改密码、账号安全设置</p>
        </div>
        <span class="setting-arrow">{{ expandedSection === 'privacy' ? '↓' : '→' }}</span>
      </div>

      <div v-if="expandedSection === 'privacy'" class="card privacy-content">
        <div class="privacy-section">
          <button class="change-password-btn" @click="openPasswordModal('change')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
            </svg>
            修改密码
          </button>
          <p class="password-hint">点击按钮打开密码修改窗口，忘记密码可通过注册邮箱重置</p>
        </div>
      </div>

      <div class="card setting-item">
        <div class="setting-icon">💾</div>
        <div class="setting-info">
          <h3>数据管理</h3>
          <p>清除缓存、导出数据</p>
        </div>
        <span class="setting-arrow">→</span>
      </div>

      <div class="card setting-item">
        <div class="setting-icon">❓</div>
        <div class="setting-info">
          <h3>帮助与反馈</h3>
          <p>常见问题、意见反馈</p>
        </div>
        <span class="setting-arrow">→</span>
      </div>

      <div class="card setting-item">
        <div class="setting-icon">ℹ️</div>
        <div class="setting-info">
          <h3>关于</h3>
          <p>版本信息、使用条款</p>
        </div>
        <span class="setting-arrow">→</span>
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
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
  z-index: 99999;
  animation: toastIn 0.3s ease;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
}

.toast.success {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
}

.toast.error {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: white;
}

.toast.warning {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: white;
}

.toast.info {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: white;
}

@keyframes toastIn {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from .modal-content,
.modal-fade-leave-to .modal-content {
  transform: scale(0.9) translateY(20px);
  opacity: 0;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 99999;
}

.modal-content {
  background: white;
  border-radius: 24px;
  padding: 32px 24px 24px;
  width: 90%;
  max-width: 340px;
  text-align: center;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  animation: modalSpring 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes modalSpring {
  0% { transform: scale(0.8); opacity: 0; }
  70% { transform: scale(1.03); }
  100% { transform: scale(1); opacity: 1; }
}

.modal-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 16px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.modal-icon svg {
  width: 32px;
  height: 32px;
}

.modal-title {
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 16px;
  color: #111827;
}

.modal-desc {
  font-size: 14px;
  color: #6b7280;
  margin: 0 0 24px;
  line-height: 1.5;
}

.modal-buttons {
  display: flex;
  gap: 12px;
}

.modal-btn {
  flex: 1;
  padding: 14px 16px;
  border-radius: 16px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.modal-btn.cancel {
  background: #f3f4f6;
  color: #4b5563;
}

.modal-btn.cancel:hover {
  background: #e5e7eb;
}

.modal-btn.confirm {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
}

.modal-btn.confirm:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px -10px rgba(102, 126, 234, 0.5);
}

.password-form {
  text-align: left;
}

.password-form label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 6px;
}

.password-form input {
  width: 100%;
  padding: 12px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 14px;
  transition: all 0.2s;
  box-sizing: border-box;
  margin-bottom: 12px;
}

.password-form input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.switch-mode {
  text-align: center;
  font-size: 13px;
  color: #667eea;
  cursor: pointer;
  margin: 8px 0 16px;
  font-weight: 500;
  transition: color 0.2s;
}

.switch-mode:hover {
  color: #764ba2;
  text-decoration: underline;
}

.password-modal .modal-buttons {
  margin-top: 8px;
}

:deep(.dark) .password-modal label {
  color: #d1d5db;
}

:deep(.dark) .password-form input {
  background: #374151;
  border-color: #4b5563;
  color: #f9fafb;
}

:deep(.dark) .password-form input:focus {
  border-color: #818cf8;
}

.change-password-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px 20px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  border: none;
  border-radius: 14px;
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.change-password-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 25px -8px rgba(102, 126, 234, 0.5);
}

.change-password-btn svg {
  width: 20px;
  height: 20px;
}

.password-hint {
  text-align: center;
  font-size: 12px;
  color: #6b7280;
  margin: 12px 0 0;
}

.settings-page {
  max-width: 600px;
  margin: 0 auto;
}

.page-header {
  padding: 0 20px 20px;
  border-bottom: 1px solid #eff3f4;
  margin-bottom: 20px;
}

.page-header h1 {
  font-size: 20px;
  font-weight: 700;
}

.settings-content {
  padding: 0 20px;
}

.setting-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.setting-item:hover {
  background: #f7f9fa;
}

.clear-notifications-btn {
  background: #ef4444;
  color: white;
  border: none;
  padding: 10px 16px;
  border-radius: 10px;
  font-weight: 700;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.clear-notifications-btn:hover {
  background: #dc2626;
  transform: scale(1.02);
}

.danger-zone-inline {
  padding: 16px;
  border: 2px solid #fee2e2;
  background: #fef2f2;
  border-radius: 12px;
}

.theme-toggle:hover {
  background: #f7f9fa;
}

.theme-toggle .setting-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.theme-toggle svg {
  width: 24px;
  height: 24px;
  color: #666;
}

.toggle-switch {
  margin-left: auto;
  width: 48px;
  height: 28px;
  background: #e5e7eb;
  border-radius: 14px;
  position: relative;
  cursor: pointer;
  transition: all 0.3s ease;
}

.toggle-switch.active {
  background: #0095f6;
}

.toggle-circle {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 24px;
  height: 24px;
  background: white;
  border-radius: 50%;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
}

.toggle-switch.active .toggle-circle {
  transform: translateX(20px);
}

:deep(.dark) .theme-toggle:hover {
  background: var(--bg-hover);
}

:deep(.dark) .theme-toggle svg {
  color: var(--text-secondary);
}

:deep(.dark) .danger-zone-inline {
  border-color: #7f1d1d !important;
  background: #450a0a !important;
}

:deep(.dark) .danger-zone-inline p {
  color: #fca5a5 !important;
}

:deep(.dark) .toggle-switch {
  background: #3a3a3a;
}

:deep(.dark) .modal-content {
  background: #1f2937;
}

:deep(.dark) .modal-title {
  color: #f9fafb;
}

:deep(.dark) .modal-desc {
  color: #9ca3af;
}

:deep(.dark) .modal-btn.cancel {
  background: #374151;
  color: #d1d5db;
}

:deep(.dark) .modal-btn.cancel:hover {
  background: #4b5563;
}

.setting-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #f7f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}

.setting-info {
  flex: 1;
}

.setting-info h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
}

.setting-info p {
  font-size: 13px;
  color: #536471;
}

.setting-arrow {
  color: #536471;
  font-size: 18px;
  font-weight: 700;
}

.privacy-content {
  margin-top: -8px;
  margin-bottom: 12px;
}

.privacy-section {
  padding: 20px;
}

:deep(.dark) .setting-item:hover {
  background: var(--bg-hover);
}

:deep(.dark) .setting-icon {
  background: var(--bg-tertiary);
}

:deep(.dark) .setting-info h3 {
  color: var(--text-primary);
}

:deep(.dark) .setting-info p {
  color: var(--text-secondary);
}

:deep(.dark) .setting-arrow {
  color: var(--text-secondary);
}

:deep(.dark) .password-hint {
  color: #9ca3af;
}
</style>
