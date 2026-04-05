<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { getConversations } from '../api/message'
import { getUserInfo } from '../api/auth'

const router = useRouter()
const userStore = useUserStore()

const conversations = ref([])
const loading = ref(false)

const fetchConversations = async () => {
  loading.value = true
  try {
    const res = await getConversations()
    const convList = res.conversations || []
    
    for (const conv of convList) {
      try {
        const userRes = await getUserInfo({ userid: conv.other_user_id })
        conv.otherUser = userRes.user
      } catch (e) {
        conv.otherUser = { username: '用户', userid: conv.other_user_id }
      }
    }
    
    conversations.value = convList
  } catch (err) {
    console.error('获取对话列表失败:', err)
  } finally {
    loading.value = false
  }
}

const goToChat = (convId) => {
  router.push(`/chat/${convId}`)
}

const formatTime = (time) => {
  if (!time) return ''
  const d = new Date(time)
  const now = new Date()
  const diff = now - d
  
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return d.toLocaleDateString('zh-CN')
}

onMounted(() => {
  fetchConversations()
})
</script>

<template>
  <div class="container">
    <h2 style="margin-bottom: 24px; font-size: 24px; font-weight: 600;">私信</h2>

    <div class="card" style="max-width: 600px; margin: 0 auto;">
      <div v-if="loading" style="padding: 60px 20px; text-align: center; color: #8e8e8e;">
        加载中...
      </div>

      <div v-else-if="conversations.length === 0" style="padding: 60px 20px; text-align: center; color: #8e8e8e;">
        <p style="font-size: 48px; margin-bottom: 16px;">💬</p>
        <p>暂无私信对话</p>
        <p style="font-size: 14px; margin-top: 8px;">去他人主页点击「私信」开始聊天吧！</p>
      </div>

      <div v-else>
        <div
          v-for="conv in conversations"
          :key="conv.conversation_id"
          class="conversation-item"
          @click="goToChat(conv.conversation_id)"
        >
          <div class="conv-avatar">
            {{ conv.otherUser?.username?.charAt(0)?.toUpperCase() || 'U' }}
          </div>
          <div class="conv-info">
            <div class="conv-header">
              <span class="conv-name">{{ conv.otherUser?.username || '用户' }}</span>
              <span class="conv-time">{{ formatTime(conv.last_message_at || conv.updated_at) }}</span>
            </div>
            <p class="conv-last-msg">
              {{ conv.last_message || '暂无消息' }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.conversation-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #efefef;
  cursor: pointer;
  transition: background 0.2s;
}

.conversation-item:hover {
  background: #f7f9fa;
}

.conv-avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea, #764ba2);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 18px;
  margin-right: 14px;
  flex-shrink: 0;
}

.conv-info {
  flex: 1;
  min-width: 0;
}

.conv-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.conv-name {
  font-weight: 700;
  font-size: 15px;
}

.conv-time {
  font-size: 12px;
  color: #8e8e8e;
}

.conv-last-msg {
  margin: 0;
  font-size: 14px;
  color: #536471;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
