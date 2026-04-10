<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useNotificationStore } from '../stores/notification'
import { getConversationMessages, sendMessage, markConversationAsRead } from '../api/message'
import { getUserInfo } from '../api/auth'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const notificationStore = useNotificationStore()

const messages = ref([])
const newMessage = ref('')
const loading = ref(false)
const targetUser = ref(null)
const messagesEndRef = ref(null)

let wsListener = null

const scrollToBottom = async () => {
  await nextTick()
  messagesEndRef.value?.scrollIntoView({ behavior: 'smooth' })
}

const fetchMessages = async () => {
  const convId = route.params.conv_id
  if (!convId || convId === 'undefined') {
    console.warn('对话ID无效，返回首页:', convId)
    router.push('/')
    return
  }
  
  loading.value = true
  try {
    const res = await getConversationMessages(convId)
    console.log('后端返回消息数量:', res.messages?.length || 0, '条')
    messages.value = res.messages || []
    
    if (res.other_user_id) {
      const userRes = await getUserInfo({ userid: res.other_user_id })
      targetUser.value = userRes.user
    }
    
    markConversationAsRead(convId)
    notificationStore.clearConversationUnread(Number(convId))
  } catch (err) {
    console.error('获取消息失败:', err)
    router.push('/')
  } finally {
    loading.value = false
    scrollToBottom() // 无动画直接跳到最底部，用户完全感知不到
  }
}

const handleSendMessage = async () => {
  if (!newMessage.value.trim()) return
  
  const content = newMessage.value.trim()
  const convId = route.params.conv_id
  
  const tempMsg = {
    id: Date.now(),
    sender_id: userStore.user.userid,
    receiver_id: targetUser.value?.userid,
    content,
    created_at: new Date().toISOString(),
    isTemp: true
  }
  
  messages.value.push(tempMsg)
  newMessage.value = ''
  scrollToBottom()
  
  try {
    await sendMessage(targetUser.value.userid, convId, content)
  } catch (err) {
    console.error('发送失败:', err)
    messages.value = messages.value.filter(m => m.id !== tempMsg.id)
  }
}

const isOwnMessage = (msg) => {
  return msg.sender_id === userStore.user?.userid
}

const formatTime = (time) => {
  if (!time) return ''
  const timestamp = typeof time === 'number' ? time * 1000 : time
  return new Date(timestamp).toLocaleString('zh-CN', { 
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

const goToUserProfile = () => {
  if (targetUser.value?.userid) {
    router.push(`/profile/${targetUser.value.userid}`)
  }
}

watch(() => route.params.conv_id, (newConvId) => {
  if (newConvId && newConvId !== 'undefined') {
    fetchMessages()
  }
}, { immediate: false })

wsListener = (event) => {
  try {
    const data = JSON.parse(event.data)
    if (data.content_type === 'private_message' && data.sender_id === targetUser.value?.userid) {
      messages.value.push({
        message_id: data.message_id,
        sender_id: data.sender_id,
        receiver_id: userStore.user?.userid,
        content: data.content,
        created_at: data.created_at
      })
      scrollToBottom()
      
      markConversationAsRead(route.params.conv_id)
      notificationStore.clearConversationUnread(Number(route.params.conv_id))
    }
  } catch (e) {}
}

onMounted(() => {
  document.body.style.overflow = 'hidden'
  
  setTimeout(() => {
    const convId = route.params.conv_id
    if (convId && convId !== 'undefined') {
      fetchMessages()
    }
  }, 50)
  
  if (notificationStore.ws) {
    notificationStore.ws.addEventListener('message', wsListener)
  }
})

onUnmounted(() => {
  document.body.style.overflow = ''
  
  if (notificationStore.ws && wsListener) {
    notificationStore.ws.removeEventListener('message', wsListener)
  }
})
</script>

<template>
  <div class="chat-page">
    <div class="chat-header">
      <button class="back-btn" @click="router.back()">←</button>
      <div class="chat-user-info" @click="goToUserProfile">
        <div class="user-avatar">
          <img v-if="targetUser?.avatarUrl" :src="targetUser.avatarUrl" alt="avatar">
          <span v-else>{{ targetUser?.username?.charAt(0)?.toUpperCase() || 'U' }}</span>
        </div>
        <span class="username">{{ targetUser?.username || '用户' }}</span>
      </div>
    </div>

    <div class="chat-messages">
      <div v-if="loading" class="loading">加载中...</div>
      
      <div
        v-for="msg in messages"
        :key="msg.id"
        class="message-item"
        :class="{ own: isOwnMessage(msg) }"
      >
        <div class="msg-avatar-chat" :class="{ own: isOwnMessage(msg) }">
          <img 
            v-if="isOwnMessage(msg) ? userStore.user?.avatarUrl : targetUser?.avatarUrl" 
            :src="isOwnMessage(msg) ? userStore.user?.avatarUrl : targetUser?.avatarUrl"
            alt="avatar"
          >
          <span v-else>
            {{ isOwnMessage(msg) 
              ? userStore.user?.username?.charAt(0)?.toUpperCase() || 'U' 
              : targetUser?.username?.charAt(0)?.toUpperCase() || 'U' 
            }}
          </span>
        </div>
        <div class="message-wrapper">
          <div class="message-content">
            {{ msg.content }}
          </div>
          <div class="message-time">{{ formatTime(msg.created_at || msg.time) }}</div>
        </div>
      </div>
      
      <div ref="messagesEndRef"></div>
    </div>

    <div class="chat-input-area">
      <input
        v-model="newMessage"
        type="text"
        placeholder="输入消息..."
        @keyup.enter="handleSendMessage"
      >
      <button @click="handleSendMessage" :disabled="!newMessage.trim()">
        发送
      </button>
    </div>
  </div>
</template>

<style scoped>
.chat-page {
  display: flex;
  flex-direction: column;
  width: calc(100vw - 575px);
  max-width: 825px;
  height: calc(100vh - 10px);
  position: fixed;
  top: 5px;
  left: 319px;
  background: #f7f9fa;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  z-index: 100;
}

.chat-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: white;
  border-bottom: 1px solid #eff3f4;
  position: sticky;
  top: 0;
  z-index: 100;
}

.back-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}

.back-btn:hover {
  background: #f7f9fa;
}

.chat-user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 12px;
  margin-left: -8px;
  transition: all 0.2s ease;
  width: fit-content;
}

.chat-user-info:hover {
  background: rgba(131, 58, 180, 0.08);
}

.chat-user-info:active {
  transform: scale(0.98);
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #833ab4 0%, #fd1d1d 50%, #fcb045 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 16px;
  overflow: hidden;
  padding: 2px;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid white;
  background: white;
}

.username {
  font-weight: 700;
  font-size: 16px;
  background: linear-gradient(135deg, #833ab4 0%, #fd1d1d 50%, #fcb045 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.loading {
  text-align: center;
  color: #536471;
  padding: 20px;
}

.msg-avatar-chat {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #833ab4 0%, #fd1d1d 50%, #fcb045 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 14px;
  flex-shrink: 0;
  padding: 2px;
  overflow: hidden;
}

.msg-avatar-chat img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid white;
  background: white;
}

.msg-avatar-chat.own {
  order: 2;
}

.message-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  width: 100%;
  justify-content: flex-start;
}

.message-item.own {
  justify-content: flex-end;
}

.message-item.own .message-wrapper {
  align-items: flex-end;
}

.message-wrapper {
  display: flex;
  flex-direction: column;
}

.message-content {
  padding: 12px 18px;
  border-radius: 22px;
  background: white;
  font-size: 15px;
  line-height: 1.4;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  max-width: 280px;
  width: fit-content;
  word-break: break-word;
  transition: transform 0.1s ease;
}

.message-item.own .message-content {
  background: linear-gradient(135deg, #833ab4 0%, #fd1d1d 50%, #fcb045 100%);
  color: white;
  box-shadow: 0 4px 15px rgba(131, 58, 180, 0.3);
}

.message-time {
  font-size: 11px;
  color: #536471;
  margin-top: 4px;
  padding: 0 4px;
}

.chat-input-area {
  display: flex;
  padding: 16px;
  background: rgba(255, 255, 255, 0.95);
  border-top: 1px solid #efefef;
  gap: 12px;
  backdrop-filter: blur(10px);
}

.chat-input-area input {
  flex: 1;
  padding: 14px 20px;
  border: 1px solid #efefef;
  border-radius: 24px;
  font-size: 15px;
  outline: none;
  background: #fafafa;
  transition: all 0.2s ease;
}

.chat-input-area input:focus {
  border-color: #0095f6;
  background: white;
  box-shadow: 0 0 0 3px rgba(0, 149, 246, 0.08);
}

.chat-input-area button {
  padding: 14px 24px;
  border-radius: 24px;
  border: none;
  background: linear-gradient(135deg, #833ab4 0%, #fd1d1d 50%, #fcb045 100%);
  color: white;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  transform: scale(1);
  box-shadow: 0 4px 15px rgba(131, 58, 180, 0.3);
}

.chat-input-area button:disabled {
  background: #dbdbdb;
  box-shadow: none;
  cursor: not-allowed;
  transform: scale(0.98);
}

.chat-input-area button:hover:not(:disabled) {
  transform: scale(0.97);
  box-shadow: 0 6px 25px rgba(253, 29, 29, 0.35);
}

.chat-input-area button:active:not(:disabled) {
  transform: scale(0.95);
}
</style>
