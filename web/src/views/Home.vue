<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useLikeStore } from '../stores/like'
import { useModalStore } from '../stores/modal'
import { likePost, unlikePost, getPostLikeCount } from '../api/like'
import { getUserPosts, getFollowingPosts, deletePost } from '../api/post'
import { getPostComments } from '../api/comment'

const modalStore = useModalStore()
const likeStore = useLikeStore()

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

const router = useRouter()
const userStore = useUserStore()

const goToProfile = (userid, event) => {
  event.stopPropagation()
  router.push(`/profile/${userid}`)
}

const posts = ref([])
const activeTab = ref('following')
const loading = ref(false)

const fetchPosts = async () => {
  loading.value = true
  try {
    let res
    if (activeTab.value === 'following') {
      res = await getFollowingPosts()
    } else {
      res = { posts: [] }
    }
    const postList = res.posts || []
    
    // 为每个帖子获取最新的点赞状态
    for (const post of postList) {
      try {
        const likeRes = await getPostLikeCount(post.postid)
        post.likeCount = likeRes.likeCount
        post.isLiked = likeRes.isLiked
        // 获取评论数
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

const handleLike = async (post, index) => {
  if (!userStore.isLoggedIn) {
    router.push('/login')
    return
  }
  try {
    if (likeStore.isLiked(post.postid)) {
      await unlikePost(post.postid)
      likeStore.removeLike(post.postid)
      posts.value[index].likeCount--
    } else {
      await likePost(post.postid)
      likeStore.addLike(post.postid)
      posts.value[index].likeCount++
    }
  } catch (err) {
    console.error('点赞失败:', err.response?.data || err)
    modalStore.openAlert('操作失败: ' + (err.response?.data?.error || '请稍后重试'), 'error')
  }
}

const goToPostDetail = (postid) => {
  router.push(`/post/${postid}`)
}

const handleDeletePost = async (postid, index) => {
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
    if (userStore.user?.avatarUrl) {
      return { backgroundImage: `url(${userStore.user.avatarUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' }
    }
    if (userStore.user?.avatarColor) {
      return { background: userStore.user.avatarColor }
    }
  }
  const hash = post.userid?.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) || 0
  return { background: avatarColors[hash % avatarColors.length] }
}

const handlePostCreated = () => {
  location.reload()
}

onMounted(() => {
  fetchPosts()
  window.addEventListener('postCreated', handlePostCreated)
})

onUnmounted(() => {
  window.removeEventListener('postCreated', handlePostCreated)
})
</script>

<template>
  <div class="container">
    <div class="home-tabs">
      <button 
        class="tab-btn" 
        :class="{ active: activeTab === 'recommend' }"
        @click="activeTab = 'recommend'; fetchPosts()"
      >
        推荐
      </button>
      <button 
        class="tab-btn" 
        :class="{ active: activeTab === 'following' }"
        @click="activeTab = 'following'; fetchPosts()"
      >
        我的关注
      </button>
    </div>

    <div v-if="activeTab === 'recommend'" style="text-align: center; padding: 60px 20px;">
      <p style="font-size: 48px; margin-bottom: 16px;">🚧</p>
      <h2 style="margin-bottom: 8px;">功能开发中</h2>
      <p style="color: #8e8e8e;">推荐内容即将上线，敬请期待！</p>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="spinner"></div>
      <p style="color: #8e8e8e; margin-top: 12px;">加载中...</p>
    </div>

    <div v-else-if="activeTab === 'following' && posts.length === 0" style="text-align: center; padding: 60px 20px;">
      <p style="font-size: 48px; margin-bottom: 16px;">📝</p>
      <h2 style="margin-bottom: 8px;">还没有帖子</h2>
      <p style="color: #8e8e8e;">点击左栏「发布帖子」按钮分享你的想法吧！</p>
    </div>

    <div v-else-if="activeTab === 'following'" v-for="(post, index) in posts" :key="post.postid" class="card post-card">
      <div class="post-header">
        <div class="post-avatar" :style="getAvatarStyle(post)" @click="goToProfile(post.userid, $event)">{{ post.avatarUrl ? '' : (post.username?.charAt(0)?.toUpperCase() || 'U') }}</div>
        <div class="post-info">
          <div class="post-username" @click="goToProfile(post.userid, $event)">{{ post.username || '用户' }}</div>
          <div class="post-time">{{ formatTime(post.posttime) }}</div>
        </div>
        <button
          v-if="post.userid === userStore.user?.userid"
          class="delete-btn"
          @click.stop="handleDeletePost(post.postid, index)"
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

      <div class="post-actions">
        <button class="action-btn" @click="handleLike(post, index)" :class="{ liked: likeStore.isLiked(post.postid) }">
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
</template>

<style scoped>
.home-tabs {
  display: flex;
  background: white;
  border-radius: 12px;
  padding: 4px;
  margin-bottom: 16px;
  gap: 4px;
}

.tab-btn {
  flex: 1;
  padding: 12px 16px;
  border: none;
  background: transparent;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 500;
  color: #536471;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  background: #f7f9fa;
}

.tab-btn.active {
  background: #1d9bf0;
  color: white;
  font-weight: 600;
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
</style>
