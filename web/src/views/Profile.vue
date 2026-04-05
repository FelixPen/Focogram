<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useLikeStore } from '../stores/like'
import { useModalStore } from '../stores/modal'
import { getUserPosts, getUserLikedPosts, deletePost } from '../api/post'
import { getPostLikeCount, likePost, unlikePost } from '../api/like'
import { getPostComments } from '../api/comment'
import { getUserInfo as fetchUserInfo, updateUserInfo, uploadAvatar } from '../api/auth'
import { followUser, unfollowUser, checkFollow } from '../api/follow'

const modalStore = useModalStore()

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const likeStore = useLikeStore()

const profileUser = ref(null)
const posts = ref([])
const likedPosts = ref([])
const isFollowing = ref(false)
const showEditModal = ref(false)
const activeTab = ref('posts')
const loading = ref(false)
const uploadingAvatar = ref(false)
const bannerColors = [
  'linear-gradient(135deg, #667eea, #764ba2)',
  'linear-gradient(135deg, #f093fb, #f5576c)',
  'linear-gradient(135deg, #4facfe, #00f2fe)',
  'linear-gradient(135deg, #43e97b, #38f9d7)',
  'linear-gradient(135deg, #fa709a, #fee140)',
  'linear-gradient(135deg, #a8edea, #fed6e3)',
  'linear-gradient(135deg, #ff9a9e, #fecfef)',
  'linear-gradient(135deg, #ffecd2, #fcb69f)',
  'linear-gradient(135deg, #a1c4fd, #c2e9fb)',
  'linear-gradient(135deg, #d299c2, #fef9d7)',
  'linear-gradient(135deg, #89f7fe, #66a6ff)',
  '#1d9bf0',
  '#14171a',
  '#ff6b6b',
  '#4ecdc4',
  '#45b7d1'
]

const editForm = ref({
  username: '',
  describe: '',
  gender: '',
  avatarColor: '',
  avatarUrl: '',
  bannerColor: '',
  birthDate: ''
})

const calculateAge = (birthDate) => {
  if (!birthDate) return null
  const today = new Date()
  const birth = new Date(birthDate)
  let age = today.getFullYear() - birth.getFullYear()
  const monthDiff = today.getMonth() - birth.getMonth()
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
    age--
  }
  return age
}

const avatarColors = [
  'linear-gradient(45deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888)',
  'linear-gradient(135deg, #667eea, #764ba2)',
  'linear-gradient(135deg, #f093fb, #f5576c)',
  'linear-gradient(135deg, #4facfe, #00f2fe)',
  'linear-gradient(135deg, #43e97b, #38f9d7)',
  'linear-gradient(135deg, #fa709a, #fee140)',
  'linear-gradient(135deg, #a8edea, #fed6e3)',
  'linear-gradient(135deg, #ff9a9e, #fecfef)',
  'linear-gradient(135deg, #ffecd2, #fcb69f)',
  'linear-gradient(135deg, #a1c4fd, #c2e9fb)',
  'linear-gradient(135deg, #d299c2, #fef9d7)',
  'linear-gradient(135deg, #89f7fe, #66a6ff)'
]

const targetUserId = computed(() => route.params.userid || userStore.user?.userid)
const isOwnProfile = computed(() => targetUserId.value === userStore.user?.userid)

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const t = new Date(timeStr)
  return t.toLocaleString('zh-CN', { 
    year: 'numeric', 
    month: '2-digit', 
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).replace(/\//g, '-')
}

const getAvatarStyle = (post) => {
  if (post.avatarUrl) {
    return { backgroundImage: `url(${post.avatarUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' }
  }
  if (post.avatarColor) {
    return { background: post.avatarColor }
  }
  if (post.userid === userStore.user?.userid) {
    if (profileUser.value?.avatarUrl) {
      return { backgroundImage: `url(${profileUser.value.avatarUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' }
    }
    if (profileUser.value?.avatarColor) {
      return { background: profileUser.value.avatarColor }
    }
  }
  const hash = post.userid?.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) || 0
  return { background: avatarColors[hash % avatarColors.length] }
}

const fetchProfile = async () => {
  try {
    const res = await fetchUserInfo({ userid: targetUserId.value })
    profileUser.value = res.user
    
    if (!isOwnProfile.value) {
      try {
        const followRes = await checkFollow(targetUserId.value)
        isFollowing.value = followRes.isFollowing
      } catch (e) {
        isFollowing.value = false
      }
    }
  } catch (err) {
    console.error('获取用户信息失败:', err)
  }
}

const fetchPosts = async () => {
  loading.value = true
  try {
    const res = await getUserPosts(targetUserId.value)
    const postList = res.posts || []
    
    for (const post of postList) {
      try {
        const likeRes = await getPostLikeCount(post.postid)
        post.likeCount = likeRes.likeCount
        post.isLiked = likeRes.isLiked
        const commentRes = await getPostComments(post.postid)
        post.commentCount = commentRes.total || 0
      } catch (e) {
        post.likeCount = 0
        post.isLiked = false
        post.commentCount = 0
      }
    }
    
    posts.value = postList
  } catch (err) {
    console.error('获取帖子失败:', err)
  } finally {
    loading.value = false
  }
}

const fetchLikedPosts = async () => {
  loading.value = true
  try {
    const res = await getUserLikedPosts(targetUserId.value)
    const postList = res.posts || []
    
    for (const post of postList) {
      try {
        const likeRes = await getPostLikeCount(post.postid)
        post.likeCount = likeRes.likeCount
        post.isLiked = likeRes.isLiked
        const commentRes = await getPostComments(post.postid)
        post.commentCount = commentRes.total || 0
      } catch (e) {
        post.likeCount = 0
        post.isLiked = false
        post.commentCount = 0
      }
    }
    
    likedPosts.value = postList
  } catch (err) {
    console.error('获取点赞帖子失败:', err)
  } finally {
    loading.value = false
  }
}

const switchTab = (tab) => {
  activeTab.value = tab
  if (tab === 'posts') {
    fetchPosts()
  } else {
    fetchLikedPosts()
  }
}

const goToPostDetail = (postid) => {
  router.push(`/post/${postid}`)
}

const handleDeletePost = async (postid, index, tab) => {
  const confirmed = await modalStore.openConfirm('确定要删除这条帖子吗？')
  if (!confirmed) return
  
  try {
    await deletePost(postid)
    modalStore.openAlert('删除成功！', 'success')
    setTimeout(() => location.reload(), 500)
  } catch (err) {
    console.error('删除失败:', err)
    modalStore.openAlert('删除失败: ' + (err.response?.data?.error || '请稍后重试'), 'error')
  }
}

const openEditModal = () => {
  if (!isOwnProfile.value) {
    modalStore.openAlert('无权编辑他人资料！', 'warning')
    return
  }
  editForm.value = {
    username: profileUser.value?.username || userStore.user?.username || '',
    describe: profileUser.value?.describe || userStore.user?.describe || '',
    gender: profileUser.value?.gender || userStore.user?.gender || '',
    avatarColor: profileUser.value?.avatarColor || userStore.user?.avatarColor || '',
    avatarUrl: profileUser.value?.avatarUrl || userStore.user?.avatarUrl || '',
    bannerColor: profileUser.value?.bannerColor || userStore.user?.bannerColor || '',
    birthDate: profileUser.value?.birthDate || userStore.user?.birthDate || ''
  }
  showEditModal.value = true
}

const handleAvatarUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  uploadingAvatar.value = true
  try {
    const res = await uploadAvatar(file)
    editForm.value.avatarUrl = res.avatarUrl
    modalStore.openAlert('上传成功！', 'success')
  } catch (err) {
    console.error('上传失败:', err)
    modalStore.openAlert('上传失败: ' + (err.response?.data?.error || err.message || '请稍后重试'), 'error')
  } finally {
    uploadingAvatar.value = false
  }
}

const handleSaveProfile = async () => {
  if (!isOwnProfile.value) {
    modalStore.openAlert('无权修改他人信息！', 'warning')
    return
  }
  try {
    await updateUserInfo(editForm.value)
    showEditModal.value = false
    // 刷新用户信息
    fetchProfile()
    // 只在自己的主页才更新 store 里的用户信息
    if (userStore.user && isOwnProfile.value) {
      userStore.user.username = editForm.value.username
      userStore.user.describe = editForm.value.describe
      userStore.user.gender = editForm.value.gender
      userStore.user.avatarColor = editForm.value.avatarColor
      userStore.user.avatarUrl = editForm.value.avatarUrl
      userStore.user.bannerColor = editForm.value.bannerColor
      userStore.user.birthDate = editForm.value.birthDate
    }
    modalStore.openAlert('保存成功！', 'success')
  } catch (err) {
    console.error('保存失败:', err)
    modalStore.openAlert('保存失败: ' + (err.response?.data?.error || '请稍后重试'), 'error')
  }
}

const handleFollow = async () => {
  try {
    if (isFollowing.value) {
      await unfollowUser(targetUserId.value)
    } else {
      await followUser(targetUserId.value)
    }
    isFollowing.value = !isFollowing.value
  } catch (err) {
    console.error('关注操作失败:', err)
  }
}

const goToUserList = (type) => {
  router.push(`/users/${targetUserId.value}/${type}`)
}

const goToChat = async () => {
  try {
    const { createConversation } = await import('../api/message')
    const res = await createConversation({ receiver_id: targetUserId.value })
    router.push(`/chat/${res.conversation_id}`)
  } catch (err) {
    console.error('创建对话失败:', err)
  }
}

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

onMounted(() => {
  fetchProfile()
  fetchPosts()
})

watch(() => route.params.userid, (newUserId, oldUserId) => {
  if (newUserId !== oldUserId) {
    profileUser.value = null
    posts.value = []
    likedPosts.value = []
    isFollowing.value = false
    fetchProfile()
    fetchPosts()
  }
})
</script>

<template>
  <div class="profile-container">
    <div class="profile-header-bar">
      <button class="back-btn" @click="router.back()">←</button>
      <div class="header-info">
        <h1>{{ profileUser?.username || '用户' }}</h1>
        <span class="post-count">{{ posts.length }} 条帖子</span>
      </div>
    </div>

    <div v-if="profileUser" class="profile-content">
      <div class="profile-banner" :style="{ background: profileUser.bannerColor || bannerColors[0] }"></div>
      
      <div class="profile-info-section">
        <div class="profile-avatar-wrapper">
          <div v-if="profileUser.avatarUrl" class="profile-avatar profile-avatar-image" :style="{ backgroundImage: `url(${profileUser.avatarUrl})` }"></div>
          <div v-else class="profile-avatar" :style="{ background: profileUser.avatarColor || avatarColors[0] }">
            {{ profileUser.username?.charAt(0)?.toUpperCase() || 'U' }}
          </div>
        </div>

        <div v-if="!isOwnProfile" class="profile-actions">
          <button
            class="follow-btn"
            :class="{ following: isFollowing }"
            @click="handleFollow"
          >
            {{ isFollowing ? '已关注' : '关注' }}
          </button>
          <button class="message-btn" @click="goToChat">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width: 18px; height: 18px;">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path>
            </svg>
            私信
          </button>
        </div>
        <button v-else class="edit-profile-btn" @click="openEditModal">编辑个人资料</button>

        <h2 class="profile-name">{{ profileUser.username || '用户' }}</h2>
        <p class="profile-account">账号：{{ profileUser.userid }}</p>

        <p class="profile-bio" v-if="profileUser.describe">{{ profileUser.describe }}</p>

        <div class="profile-details">
          <div class="detail-item">
            <svg class="detail-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <span>{{ profileUser.gender === 'male' ? '男' : profileUser.gender === 'female' ? '女' : '未设置' }}</span>
          </div>
          <div class="detail-item" v-if="calculateAge(profileUser.birthDate)">
            <svg class="detail-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3.055 11H5a2 2 0 0 1 2 2v1a2 2 0 0 0 2 2 2 2 0 0 1 2 2v2.945"></path>
              <path d="M25 10.655V15a2 2 0 0 1-2 2v-3a2 2 0 0 0-2-2 2 2 0 0 1 2-2V11.05"></path>
              <rect x="7" y="11" width="10" height="10" rx="2"></rect>
              <circle cx="12" cy="16" r="1"></circle>
            </svg>
            <span>{{ calculateAge(profileUser.birthDate) }} 岁</span>
          </div>
          <div class="detail-item">
            <svg class="detail-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
              <line x1="16" y1="2" x2="16" y2="6"></line>
              <line x1="8" y1="2" x2="8" y2="6"></line>
              <line x1="3" y1="10" x2="21" y2="10"></line>
            </svg>
            <span>注册于 {{ profileUser.createdAt ? profileUser.createdAt.substring(0, 10) : '未知' }}</span>
          </div>
        </div>

        <div class="profile-stats">
          <span class="stat-item" :class="{ clickable: isOwnProfile }" @click="isOwnProfile && goToUserList('following')"><strong>{{ profileUser.followingCount || 0 }}</strong> 关注</span>
          <span class="stat-item" :class="{ clickable: isOwnProfile }" @click="isOwnProfile && goToUserList('followers')"><strong>{{ profileUser.followersCount || 0 }}</strong> 粉丝</span>
        </div>
      </div>

      <div class="profile-tabs">
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'posts' }"
          @click="switchTab('posts')"
        >
          帖子
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'liked' }"
          @click="switchTab('liked')"
        >
          喜欢
        </button>
      </div>

      <div class="profile-posts">
        <div v-if="loading" class="loading-container">
          <div class="spinner"></div>
          <p style="color: #8e8e8e; margin-top: 12px;">加载中...</p>
        </div>

        <div v-else-if="activeTab === 'posts' && posts.length === 0" class="empty-posts">
          <p style="font-size: 48px; margin-bottom: 16px;">📝</p>
          <p>还没有发布任何帖子</p>
        </div>

        <div v-else-if="activeTab === 'liked' && likedPosts.length === 0" class="empty-posts">
          <p style="font-size: 48px; margin-bottom: 16px;">❤️</p>
          <p>还没有点赞任何帖子</p>
        </div>

        <div v-if="activeTab === 'posts'" v-for="(post, index) in posts" :key="post.postid" class="card post-card" @click="goToPostDetail(post.postid)">
          <div class="post-header">
            <div class="post-avatar" :style="getAvatarStyle(post)">{{ post.avatarUrl ? '' : (post.username?.charAt(0)?.toUpperCase() || 'U') }}</div>
            <div class="post-info">
              <div class="post-username">{{ post.username || '用户' }}</div>
              <div class="post-time">{{ formatTime(post.posttime) }}</div>
            </div>
            <button 
              v-if="post.userid === userStore.user?.userid" 
              class="delete-btn"
              @click.stop="handleDeletePost(post.postid, index, 'posts')"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="3 6 5 6 21 6"></polyline>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              </svg>
            </button>
          </div>

          <div class="post-caption">
            {{ post.content }}
          </div>

        <img v-if="post.image || post.image_url" class="post-image" :src="post.image || post.image_url" alt="帖子图片">

          <div class="post-actions" @click.stop>
            <button class="action-btn" :class="{ liked: likeStore.isLiked(post.postid) }" @click="toggleLike(post)">
              <svg v-if="likeStore.isLiked(post.postid)" viewBox="0 0 24 24" fill="#ed4956" stroke="#ed4956" stroke-width="2">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
              </svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
              </svg>
              <span>{{ post.likeCount || 0 }}</span>
            </button>
            <button class="action-btn" @click="goToPostDetail(post.postid)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path>
              </svg>
              <span>{{ post.commentCount || 0 }}</span>
            </button>
          </div>
        </div>

        <div v-if="activeTab === 'liked'" v-for="(post, index) in likedPosts" :key="post.postid" class="card post-card" @click="goToPostDetail(post.postid)">
          <div class="post-header">
            <div class="post-avatar" :style="getAvatarStyle(post)">{{ post.avatarUrl ? '' : (post.username?.charAt(0)?.toUpperCase() || 'U') }}</div>
            <div class="post-info">
              <div class="post-username">{{ post.username || '用户' }}</div>
              <div class="post-time">{{ formatTime(post.posttime) }}</div>
            </div>
            <button 
              v-if="post.userid === userStore.user?.userid" 
              class="delete-btn"
              @click.stop="handleDeletePost(post.postid, index, 'liked')"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="3 6 5 6 21 6"></polyline>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              </svg>
            </button>
          </div>

          <div v-if="post.image || post.image_url" class="post-image" :style="{ backgroundImage: `url(${post.image || post.image_url})` }"></div>

          <div class="post-caption">
            {{ post.content }}
          </div>

          <div class="post-actions" @click.stop>
            <button class="action-btn" :class="{ liked: likeStore.isLiked(post.postid) }" @click="toggleLike(post)">
              <svg v-if="likeStore.isLiked(post.postid)" viewBox="0 0 24 24" fill="#ed4956" stroke="#ed4956" stroke-width="2">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
              </svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
              </svg>
              <span>{{ post.likeCount || 0 }}</span>
            </button>
            <button class="action-btn" @click="goToPostDetail(post.postid)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path>
              </svg>
              <span>{{ post.commentCount || 0 }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div class="edit-modal">
        <div class="modal-header">
          <h2>编辑个人资料</h2>
          <button class="modal-close" @click="showEditModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>上传头像</label>
            <div class="avatar-upload-area">
              <div v-if="editForm.avatarUrl" class="uploaded-avatar">
                <img :src="editForm.avatarUrl" alt="avatar">
                <button class="remove-avatar-btn" @click="editForm.avatarUrl = ''">×</button>
              </div>
              <label v-else class="upload-label">
                <input type="file" accept="image/*" @change="handleAvatarUpload" style="display: none;">
                <div class="upload-placeholder">
                  <span v-if="uploadingAvatar">上传中...</span>
                  <span v-else>📷 点击上传图片</span>
                </div>
              </label>
            </div>
          </div>

          <div class="form-group">
            <label>主页横幅背景</label>
            <div class="banner-picker">
              <div
                v-for="(color, index) in bannerColors"
                :key="index"
                class="banner-option"
                :class="{ selected: editForm.bannerColor === color || (!editForm.bannerColor && index === 0) }"
                :style="{ background: color }"
                @click="editForm.bannerColor = color"
              ></div>
            </div>
          </div>

          <div class="form-group">
            <label>用户名</label>
            <input v-model="editForm.username" type="text" placeholder="请输入用户名">
          </div>
          <div class="form-group">
            <label>个人简介</label>
            <textarea v-model="editForm.describe" rows="3" placeholder="介绍一下自己吧..." maxlength="100"></textarea>
          </div>
          <div class="form-group">
            <label>性别</label>
            <select v-model="editForm.gender">
              <option value="">未设置</option>
              <option value="male">男</option>
              <option value="female">女</option>
            </select>
          </div>
          <div class="form-group">
            <label>出生日期</label>
            <input v-model="editForm.birthDate" type="date">
          </div>
        </div>
        <div class="modal-footer">
          <button class="save-btn" @click="handleSaveProfile">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-container {
  max-width: 600px;
  margin: 0 auto;
  border-left: 1px solid #eff3f4;
  border-right: 1px solid #eff3f4;
  min-height: 100vh;
}

.profile-header-bar {
  position: sticky;
  top: 0;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 8px 16px;
  border-bottom: 1px solid #eff3f4;
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
  transition: background 0.2s;
}

.back-btn:hover {
  background: #f7f9fa;
}

.header-info h1 {
  font-size: 20px;
  font-weight: 800;
  margin: 0;
}

.header-info .post-count {
  font-size: 13px;
  color: #536471;
}

.profile-banner {
  height: 150px;
}

.profile-info-section {
  padding: 12px 16px;
  position: relative;
}

.profile-avatar-wrapper {
  position: absolute;
  top: -50px;
}

.profile-avatar {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: linear-gradient(45deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 40px;
  font-weight: 600;
  border: 4px solid white;
}

.profile-actions {
  float: right;
  display: flex;
  gap: 10px;
  margin-top: 50px;
}

.edit-profile-btn {
  float: right;
  margin-top: 50px;
}

.follow-btn,
.message-btn,
.edit-profile-btn {
  padding: 8px 20px;
  border-radius: 9999px;
  font-weight: 700;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.follow-btn {
  border: 1px solid #1d9bf0;
  background: #1d9bf0;
  color: white;
}

.follow-btn.following {
  border: 1px solid #cfd9de;
  background: transparent;
  color: #0f1419;
}

.follow-btn:hover {
  background: #1a8cd8;
}

.follow-btn.following:hover {
  border-color: #f4212e;
  color: #f4212e;
  background: rgba(244, 33, 46, 0.1);
}

.edit-profile-btn {
  border: 1px solid #cfd9de;
  background: transparent;
  color: #0f1419;
}

.edit-profile-btn:hover {
  background: #f7f9fa;
}

.message-btn {
  border: 1px solid #1d9bf0;
  background: transparent;
  color: #1d9bf0;
}

.message-btn:hover {
  background: rgba(29, 155, 240, 0.1);
}

.profile-name {
  margin: 56px 0 2px 0;
  font-size: 20px;
  font-weight: 800;
}

.profile-account {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #536471;
}

.profile-bio {
  margin: 0 0 12px 0;
  font-size: 15px;
  line-height: 20px;
  color: #0f1419;
}

.profile-details {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 12px;
  font-size: 14px;
  color: #536471;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.detail-icon {
  width: 16px;
  height: 16px;
  opacity: 0.7;
  flex-shrink: 0;
}

:deep(.dark) .detail-icon {
  opacity: 0.8;
}

.profile-stats {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #536471;
}

.stat-item {
  transition: opacity 0.2s;
}

.stat-item.clickable {
  cursor: pointer;
}

.stat-item.clickable:hover {
  opacity: 0.7;
}

.stat-item strong {
  color: #0f1419;
  font-weight: 700;
}

.profile-tabs {
  display: flex;
  border-bottom: 1px solid #eff3f4;
}

.tab-btn {
  flex: 1;
  padding: 16px;
  border: none;
  background: transparent;
  font-size: 15px;
  font-weight: 500;
  color: #536471;
  cursor: pointer;
  position: relative;
  transition: background 0.2s;
}

.tab-btn:hover {
  background: #f7f9fa;
}

.tab-btn.active {
  color: #0f1419;
  font-weight: 700;
}

.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 56px;
  height: 4px;
  background: #1d9bf0;
  border-radius: 9999px;
}

.profile-posts {
  min-height: 300px;
}

.empty-posts {
  text-align: center;
  padding: 60px 20px;
  color: #536471;
}

.post-card {
  cursor: pointer;
  transition: background 0.2s;
}

.post-card:hover {
  background: #f7f9fa;
}

.profile-bio {
  margin: 0 0 12px 0;
  font-size: 15px;
  line-height: 20px;
  color: #0f1419;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.edit-modal {
  background: white;
  border-radius: 16px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #eff3f4;
}

.modal-header h2 {
  font-size: 20px;
  font-weight: 800;
  margin: 0;
}

.modal-close {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-close:hover {
  background: #f7f9fa;
}

.modal-body {
  padding: 20px;
}

.modal-footer {
  padding: 16px 20px;
  border-top: 1px solid #eff3f4;
  display: flex;
  justify-content: flex-end;
}

.save-btn {
  padding: 10px 24px;
  border-radius: 9999px;
  border: none;
  background: #1d9bf0;
  color: white;
  font-weight: 700;
  cursor: pointer;
  transition: background 0.2s;
}

.save-btn:hover {
  background: #1a8cd8;
}

.edit-modal .form-group {
  margin-bottom: 20px;
}

.edit-modal .form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 700;
  font-size: 14px;
  color: #0f1419;
}

.edit-modal .form-group input,
.edit-modal .form-group textarea,
.edit-modal .form-group select {
  width: 100%;
  padding: 12px;
  border: 1px solid #cfd9de;
  border-radius: 4px;
  font-size: 15px;
  box-sizing: border-box;
  transition: border-color 0.2s;
}

.edit-modal .form-group input:focus,
.edit-modal .form-group textarea:focus,
.edit-modal .form-group select:focus {
  outline: none;
  border-color: #1d9bf0;
}

.edit-modal .form-group input:disabled {
  background: #f7f9fa;
  color: #536471;
}

.edit-modal .form-group textarea {
  resize: none;
}

.avatar-picker {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.avatar-option {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 50%;
  cursor: pointer;
  border: 3px solid transparent;
  transition: transform 0.2s, border-color 0.2s;
}

.avatar-option:hover {
  transform: scale(1.1);
}

.avatar-option.selected {
  border-color: #1d9bf0;
  transform: scale(1.1);
}

.banner-picker {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 8px;
}

.banner-option {
  height: 32px;
  border-radius: 6px;
  cursor: pointer;
  border: 3px solid transparent;
  transition: transform 0.2s, border-color 0.2s;
}

.banner-option:hover {
  transform: scale(1.05);
}

.banner-option.selected {
  border-color: #1d9bf0;
  transform: scale(1.05);
}

.avatar-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f7f9fa;
  border-radius: 8px;
}

.preview-circle {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 20px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #1d9bf0;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.preview-label {
  font-size: 14px;
  color: #536471;
}

.avatar-upload-area {
  margin-bottom: 16px;
}

.uploaded-avatar {
  position: relative;
  display: inline-block;
}

.uploaded-avatar img {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid #1d9bf0;
}

.remove-avatar-btn {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: none;
  background: #ff3b30;
  color: white;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-label {
  display: block;
  cursor: pointer;
}

.upload-placeholder {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #f7f9fa;
  border: 2px dashed #cfd9de;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  font-size: 12px;
  color: #536471;
  transition: all 0.2s;
}

.upload-placeholder:hover {
  border-color: #1d9bf0;
  background: #e8f5fd;
}

.profile-avatar-image {
  background-size: cover;
  background-position: center;
  border: 3px solid white;
}
</style>
