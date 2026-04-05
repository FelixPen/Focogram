<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from './stores/user'
import { useNotificationStore } from './stores/notification'
import { useLikeStore } from './stores/like'
import { useModalStore } from './stores/modal'
import { useThemeStore } from './stores/theme'
import { createPost, uploadPostImage } from './api/post'
import { likePost, unlikePost } from './api/like'

const modalStore = useModalStore()
const themeStore = useThemeStore()

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const notificationStore = useNotificationStore()
const likeStore = useLikeStore()

const showSidebar = computed(() => userStore.isLoggedIn)
const activeRightTab = ref('notifications')
const showCreateModal = ref(false)

const notifications = computed(() => notificationStore.notifications)
const unreadCount = computed(() => notificationStore.unreadCount)
const conversations = ref([])
const unreadMessageCount = computed(() => notificationStore.unreadMessageCount)
const conversationUnreadCounts = computed(() => notificationStore.conversationUnreadCounts)

const switchToNotifications = () => {
  activeRightTab.value = 'notifications'
  notificationStore.markAsRead()
}

const switchToMessages = () => {
  activeRightTab.value = 'messages'
  fetchConversations()
}

const fetchConversations = async () => {
  try {
    const { getConversations } = await import('./api/message')
    const { getUserInfo } = await import('./api/auth')
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
  }
}

const goToChat = (convId) => {
  router.push(`/chat/${convId}`)
}
const newPost = ref({ content: '', image_url: '' })
const posting = ref(false)
const uploadingImage = ref(false)
const imagePreview = ref('')
const error = ref('')

const handleImageUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  if (!file.type.startsWith('image/')) {
    error.value = '请选择图片文件'
    return
  }

  if (file.size > 10 * 1024 * 1024) {
    error.value = '图片大小不能超过 10MB'
    return
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    imagePreview.value = e.target.result
  }
  reader.readAsDataURL(file)

  uploadingImage.value = true
  error.value = ''
  try {
    const res = await uploadPostImage(file)
    newPost.value.image_url = res.imageUrl
    error.value = ''
  } catch (err) {
    error.value = err.response?.data?.error || err?.data?.error || '图片上传失败，请稍后重试'
    imagePreview.value = ''
  } finally {
    uploadingImage.value = false
  }
}

const removeImage = () => {
  imagePreview.value = ''
  newPost.value.image_url = ''
}

const handleLogout = () => {
  userStore.logout()
  router.push('/login')
}

const goToProfile = () => {
  router.push(`/profile/${userStore.user?.userid}`)
}

const handleCreatePost = async () => {
  if (!newPost.value.content.trim() && !newPost.value.image_url) {
    error.value = '请输入内容或上传图片'
    return
  }

  posting.value = true
  error.value = ''
  try {
    const res = await createPost(newPost.value)
    console.log('发帖成功:', res)
    showCreateModal.value = false
    newPost.value = { content: '', image_url: '' }
    imagePreview.value = ''
    modalStore.openAlert('发布成功！', 'success')
    window.dispatchEvent(new CustomEvent('postCreated'))
  } catch (err) {
    console.error('发布帖子失败:', err)
    error.value = err.response?.data?.error || '发布失败，请稍后重试'
  } finally {
    posting.value = false
  }
}

onMounted(() => {
  if (userStore.isLoggedIn) {
    notificationStore.connectWebSocket()
    likeStore.fetchLikedPosts()
  }
})

watch(() => userStore.isLoggedIn, (loggedIn) => {
  if (loggedIn) {
    notificationStore.connectWebSocket()
    likeStore.fetchLikedPosts()
  } else {
    notificationStore.disconnectWebSocket()
    notificationStore.clearAll()
    likeStore.clearAll()
  }
})

const toggleLike = async (post) => {
  if (!userStore.isLoggedIn) {
    router.push('/login')
    return
  }
  try {
    if (likeStore.isLiked(post.postid)) {
      await unlikePost(post.postid)
      likeStore.removeLike(post.postid)
      post.likeCount--
      post.isLiked = false
    } else {
      await likePost(post.postid)
      likeStore.addLike(post.postid)
      post.likeCount++
      post.isLiked = true
    }
  } catch (err) {
    console.error('点赞操作失败:', err)
  }
}
</script>

<template>
  <div id="app">
    <div v-if="showSidebar" class="layout-container">
      <aside class="sidebar">
        <div class="sidebar-logo">Focogram</div>

        <nav class="sidebar-nav">
          <router-link to="/" class="nav-item">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
              <polyline points="9 22 9 12 15 12 15 22"></polyline>
            </svg>
            <span class="nav-text">首页</span>
          </router-link>
          <router-link to="/search" class="nav-item">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <span class="nav-text">搜索</span>
          </router-link>
          <button @click="showCreateModal = true" class="nav-item" style="width: 100%;">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"></line>
              <line x1="5" y1="12" x2="19" y2="12"></line>
            </svg>
            <span class="nav-text">发布帖子</span>
          </button>
          <a @click="goToProfile" class="nav-item" style="cursor: pointer">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <span class="nav-text">个人主页</span>
          </a>
          <router-link to="/settings" class="nav-item">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3"></circle>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
            </svg>
            <span class="nav-text">设置</span>
          </router-link>
          <a @click="handleLogout" class="nav-item" style="cursor: pointer">
            <svg class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
              <polyline points="16 17 21 12 16 7"></polyline>
              <line x1="21" y1="12" x2="9" y2="12"></line>
            </svg>
            <span class="nav-text">退出登录</span>
          </a>
        </nav>
      </aside>

      <main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>

      <aside class="right-sidebar">
        <div class="right-tabs">
          <button 
            class="right-tab-btn" 
            :class="{ active: activeRightTab === 'notifications' }"
            @click="switchToNotifications"
          >
            🔔 通知
            <span v-if="unreadCount > 0" class="badge">
              {{ unreadCount > 99 ? '99+' : unreadCount }}
            </span>
          </button>
          <button 
            class="right-tab-btn" 
            :class="{ active: activeRightTab === 'messages' }"
            @click="switchToMessages"
          >
            ✉️ 消息
            <span v-if="unreadMessageCount > 0" class="badge">
              {{ unreadMessageCount > 99 ? '99+' : unreadMessageCount }}
            </span>
          </button>
        </div>

        <div class="right-content">
          <div v-if="activeRightTab === 'notifications'" class="notifications-list">
            <div 
              v-for="item in notifications" 
              :key="item.id" 
              class="notification-item"
            >
              <div class="notif-icon">
                <span v-if="item.content_type === 'like'">❤️</span>
                <span v-else-if="item.content_type === 'comment'">💬</span>
                <span v-else-if="item.content_type === 'follow'">👤</span>
                <span v-else>📢</span>
              </div>
              <div class="notif-content">
                <div class="notif-text">
                  {{ item.content }}
                </div>
                <div class="notif-time">{{ item.time }}</div>
              </div>
            </div>
            <div v-if="notifications.length === 0" class="empty-state">
              暂无通知
            </div>
          </div>

          <div v-if="activeRightTab === 'messages'" class="messages-list">
            <div
              v-for="conv in conversations"
              :key="conv.conversation_id"
              class="message-item"
              @click="goToChat(conv.conversation_id)"
            >
              <div class="msg-avatar">
                <img v-if="conv.otherUser?.avatarUrl" :src="conv.otherUser.avatarUrl" alt="avatar">
                <span v-else>{{ conv.otherUser?.username?.charAt(0)?.toUpperCase() || 'U' }}</span>
              </div>
              <div class="msg-content">
                <div class="msg-header">
                  <span class="msg-user">{{ conv.otherUser?.username || '用户' }}</span>
                  <span v-if="conversationUnreadCounts[conv.conversation_id] > 0" class="msg-unread-badge">
                    {{ conversationUnreadCounts[conv.conversation_id] > 99 ? '99+' : conversationUnreadCounts[conv.conversation_id] }}
                  </span>
                </div>
                <div class="msg-text">
                  {{ conv.last_message || '暂无消息' }}
                </div>
              </div>
            </div>
            <div v-if="conversations.length === 0" class="empty-state">
              暂无对话
              <p style="font-size: 12px; margin-top: 8px;">去他人主页点击「💬 私信」开始聊天</p>
            </div>
          </div>
        </div>
      </aside>

      <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
        <div class="modal-content create-post-modal">
          <div class="modal-header">
            <h2>创建新帖子</h2>
            <button class="modal-close" @click="showCreateModal = false">×</button>
          </div>

          <div v-if="error" class="error-message">{{ error }}</div>

          <div v-if="imagePreview" class="image-preview-container">
            <img :src="imagePreview" class="post-image-preview" alt="preview">
            <button class="remove-image-btn" @click="removeImage" :disabled="uploadingImage">×</button>
            <div v-if="uploadingImage" class="uploading-overlay">
              <div class="spinner"></div>
              <span>上传中...</span>
            </div>
          </div>

          <div v-else class="upload-area">
            <input
              type="file"
              id="post-image-upload"
              accept="image/*"
              @change="handleImageUpload"
              style="display: none;"
            >
            <label for="post-image-upload" class="upload-label">
              <span class="upload-icon">📷</span>
              <span class="upload-text">点击上传图片</span>
              <span class="upload-hint">支持 JPG、PNG、GIF、WEBP，最大 10MB</span>
            </label>
          </div>

          <div class="form-group">
            <textarea
              v-model="newPost.content"
              rows="3"
              placeholder="写下你的想法..."
              style="resize: none;"
              class="post-textarea"
            ></textarea>
          </div>

          <button class="btn btn-primary" style="width: 100%;" @click="handleCreatePost" :disabled="posting || uploadingImage">
            {{ posting ? '发布中...' : '发布帖子' }}
          </button>
        </div>
      </div>

      <transition name="modal-fade">
        <div v-if="modalStore.showAlert" class="modal-overlay" @click.self="modalStore.closeAlert()">
          <div class="alert-modal">
            <div class="alert-icon">
              <span v-if="modalStore.alertType === 'success'">✅</span>
              <span v-else-if="modalStore.alertType === 'error'">❌</span>
              <span v-else-if="modalStore.alertType === 'warning'">⚠️</span>
              <span v-else>ℹ️</span>
            </div>
            <p class="alert-message">{{ modalStore.alertMessage }}</p>
            <button class="alert-btn primary" @click="modalStore.closeAlert()">确定</button>
          </div>
        </div>
      </transition>

      <transition name="modal-fade">
        <div v-if="modalStore.showConfirm" class="modal-overlay">
          <div class="confirm-modal">
            <div class="confirm-icon">❓</div>
            <p class="confirm-message">{{ modalStore.confirmMessage }}</p>
            <div class="confirm-buttons">
              <button class="confirm-btn cancel" @click="modalStore.handleConfirm(false)">取消</button>
              <button class="confirm-btn primary" @click="modalStore.handleConfirm(true)">确定</button>
            </div>
          </div>
        </div>
      </transition>
    </div>

    <main v-else>
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>
