<script setup>
import { ref, onMounted, nextTick, computed, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useLikeStore } from '../stores/like'
import { useModalStore } from '../stores/modal'
import { likePost, getPostLikeCount, unlikePost } from '../api/like'
import { createComment, getPostComments, deleteComment } from '../api/comment'
import { getPostDetail, deletePost } from '../api/post'

const likeStore = useLikeStore()

const modalStore = useModalStore()

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

const getAvatarStyle = (item) => {
  if (typeof item === 'object') {
    if (item.avatarUrl) {
      return { backgroundImage: `url(${item.avatarUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' }
    }
    if (item.avatarColor) {
      return { background: item.avatarColor }
    }
    if (item.userid === userStore.user?.userid) {
      if (userStore.user?.avatarUrl) {
        return { backgroundImage: `url(${userStore.user.avatarUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' }
      }
      if (userStore.user?.avatarColor) {
        return { background: userStore.user.avatarColor }
      }
    }
    const hash = item.userid?.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) || 0
    return { background: avatarColors[hash % avatarColors.length] }
  }
  if (item === userStore.user?.userid) {
    if (userStore.user?.avatarUrl) {
      return { backgroundImage: `url(${userStore.user.avatarUrl})`, backgroundSize: 'cover', backgroundPosition: 'center' }
    }
    if (userStore.user?.avatarColor) {
      return { background: userStore.user.avatarColor }
    }
  }
  const hash = item?.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) || 0
  return { background: avatarColors[hash % avatarColors.length] }
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const postid = route.params.postid
const post = ref(null)
const comments = ref([])
const commentContent = ref('')
const postingComment = ref(false)
const showCommentBox = ref(false)

const showContextMenu = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const selectedComment = ref(null)

const canDeleteComment = (comment) => {
  if (!userStore.user) return false
  if (comment.userid === userStore.user.userid) return true
  if (post.value?.userid === userStore.user.userid) return true
  return false
}

const handleContextMenu = (e, comment) => {
  if (!canDeleteComment(comment)) return
  e.preventDefault()
  e.stopPropagation()
  selectedComment.value = comment
  contextMenuPosition.value = { x: e.clientX, y: e.clientY }
  showContextMenu.value = true
}

const closeContextMenu = () => {
  showContextMenu.value = false
  selectedComment.value = null
}

const handleDeleteComment = async () => {
  if (!selectedComment.value) return
  const commentid = selectedComment.value.commentid
  comments.value = comments.value.filter(c => c.commentid !== commentid)
  closeContextMenu()
  try {
    await deleteComment(commentid)
  } catch (err) {
    console.log('评论已删除:', err)
  }
}

onMounted(() => {
  document.addEventListener('click', closeContextMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
})

const fetchPostDetail = async () => {
  try {
    // 直接获取帖子详情
    const postRes = await getPostDetail(postid)
    console.log('帖子详情返回:', postRes)
    post.value = postRes
    
    // 获取点赞状态
    const likeRes = await getPostLikeCount(postid)
    post.value.likeCount = likeRes.likeCount
    post.value.isLiked = likeRes.isLiked
    
    // 获取评论列表
    const commentsRes = await getPostComments(postid)
    comments.value = commentsRes.comments || commentsRes.content || []
  } catch (err) {
    console.error('加载失败:', err)
  }
}

const handleLike = async () => {
  if (!userStore.isLoggedIn) {
    router.push('/login')
    return
  }
  try {
    if (likeStore.isLiked(postid)) {
      await unlikePost(postid)
      likeStore.removeLike(postid)
      post.value.likeCount--
    } else {
      await likePost(postid)
      likeStore.addLike(postid)
      post.value.likeCount++
    }
  } catch (err) {
    console.error('点赞失败:', err)
  }
}

const handleDeletePost = async () => {
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

const toggleCommentBox = () => {
  showCommentBox.value = !showCommentBox.value
  if (showCommentBox.value) {
    nextTick(() => {
      document.querySelector('.comment-input-area')?.focus()
    })
  }
}

const handlePostComment = async () => {
  if (!commentContent.value.trim()) return

  postingComment.value = true
  try {
    await createComment(postid, { content: commentContent.value })
    commentContent.value = ''
    
    const res = await getPostComments(postid)
    comments.value = res.comments || res.content || []
    showCommentBox.value = false
  } catch (err) {
    console.error('评论失败:', err)
    modalStore.openAlert('评论失败: ' + (err.response?.data?.error || '请稍后重试'), 'error')
  } finally {
    postingComment.value = false
  }
}

const goBack = () => {
  router.back()
}

const goToProfile = (userid) => {
  router.push(`/profile/${userid}`)
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

onMounted(() => {
  fetchPostDetail()
})
</script>

<template>
  <div class="post-detail-container">
    <div class="detail-header">
      <button class="back-btn" @click="goBack">←</button>
      <h1>帖子</h1>
    </div>

    <div v-if="post" class="post-detail-card">
      <div class="post-author-section">
        <div class="author-avatar" :style="getAvatarStyle(post)" @click="goToProfile(post.userid)">{{ post.avatarUrl ? '' : (post.username?.charAt(0)?.toUpperCase() || 'U') }}</div>
        <div class="author-info">
          <div class="author-name" @click="goToProfile(post.userid)">{{ post.username || '用户' }}</div>
          <div class="author-time">{{ formatTime(post.posttime) }}</div>
        </div>
        <button 
          v-if="post.userid === userStore.user?.userid" 
          class="delete-btn"
          @click="handleDeletePost"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
        </button>
      </div>

      <div class="post-content" style="white-space: pre-wrap;">{{ post.content }}</div>

      <img v-if="post.image || post.image_url" class="post-image" :src="post.image || post.image_url" alt="帖子图片" @error="(e) => e.target.style.display = 'none'">

      <div class="post-stats">
        <span>{{ post.likeCount || 0 }} 次点赞</span>
        <span>{{ comments.length }} 条评论</span>
      </div>

      <div class="post-actions-bar">
        <button @click="handleLike" :class="{ liked: likeStore.isLiked(postid) }">
          <svg v-if="likeStore.isLiked(postid)" viewBox="0 0 24 24" fill="#ed4956" stroke="#ed4956" stroke-width="2">
            <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
          </svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
          </svg>
          <span>点赞</span>
        </button>
        <button @click="toggleCommentBox" :class="{ active: showCommentBox }">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path>
          </svg>
          <span>评论</span>
        </button>
      </div>
    </div>

    <div class="comment-section">
      <div v-if="showCommentBox" class="comment-input-area">
        <div class="current-user-avatar" :style="getAvatarStyle(userStore.user?.userid)">
          <template v-if="!userStore.user?.avatarUrl">
            {{ userStore.user?.username?.charAt(0)?.toUpperCase() || 'U' }}
          </template>
        </div>
        <div class="input-wrapper">
          <textarea
            v-model="commentContent"
            rows="3"
            placeholder="发布你的评论..."
          ></textarea>
          <button 
            class="btn btn-primary" 
            :disabled="!commentContent.trim() || postingComment"
            @click="handlePostComment"
          >
            {{ postingComment ? '发布中...' : '评论' }}
          </button>
        </div>
      </div>

      <div class="comments-list">
        <div v-if="comments.length === 0" class="empty-comments">
          还没有人评论，快来抢沙发吧～
        </div>
        <div v-for="comment in comments" :key="comment.commentid" class="comment-item" @contextmenu="handleContextMenu($event, comment)">
          <div class="comment-avatar" :style="getAvatarStyle(comment)" @click.stop="goToProfile(comment.userid)">
            <template v-if="!comment.avatarUrl && comment.userid !== userStore.user?.userid">
              {{ comment.username?.charAt(0)?.toUpperCase() || 'U' }}
            </template>
          </div>
          <div class="comment-content">
            <div class="comment-header">
              <span class="comment-author" @click.stop="goToProfile(comment.userid)">{{ comment.username || '用户' }}</span>
              <span class="comment-time">{{ comment.createdAt }}</span>
            </div>
            <div class="comment-text">{{ comment.content }}</div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showContextMenu" class="context-menu" :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }" @click.stop>
      <div class="context-menu-item delete" @click="handleDeleteComment">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
        </svg>
        <span>删除评论</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.post-detail-container {
  max-width: 600px;
  margin: 0 auto;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 12px 16px;
  border-bottom: 1px solid #eff3f4;
  position: sticky;
  top: 0;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
}

.back-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: transparent;
  font-size: 18px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-btn:hover {
  background: #f7f9f9;
}

.detail-header h1 {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
}

.post-detail-card {
  padding: 20px 16px;
  border-bottom: 1px solid #eff3f4;
}

.post-author-section {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.post-author-section .delete-btn {
  margin-left: auto;
}

.author-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 600;
  flex-shrink: 0;
}

.author-info {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.author-name {
  font-weight: 700;
  font-size: 15px;
}

.author-time {
  font-size: 13px;
  color: #536471;
}

.post-content {
  font-size: 23px;
  line-height: 28px;
  margin-bottom: 16px;
  font-weight: 400;
}

.post-image {
  width: calc(100% - 24px);
  max-height: 600px;
  object-fit: contain;
  object-position: left center;
  border-radius: 16px;
  margin-bottom: 16px;
  margin-left: 12px;
  display: block;
}

.post-stats {
  padding: 16px 0;
  border-bottom: 1px solid #eff3f4;
  color: #536471;
  font-size: 14px;
  display: flex;
  gap: 20px;
}

.post-stats span {
  cursor: pointer;
}

.post-stats span:hover {
  text-decoration: underline;
}

.post-actions-bar {
  display: flex;
  justify-content: space-around;
  padding: 8px 0;
  border-bottom: 1px solid #eff3f4;
}

.post-actions-bar button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  background: transparent;
  color: #536471;
  font-size: 14px;
  cursor: pointer;
  border-radius: 9999px;
}

.post-actions-bar button svg {
  width: 22px;
  height: 22px;
}

.post-actions-bar .action-btn:hover,
.action-btn.active {
  background: rgba(29, 155, 240, 0.1);
  color: #1d9bf0;
}

.post-actions-bar button.liked {
  color: #f91880;
}

.post-actions-bar button.liked:hover {
  background: rgba(249, 24, 128, 0.1);
}

.post-actions-bar button.active {
  color: #1d9bf0;
}

.comment-section {
  padding: 0;
}

.comment-input-area {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid #eff3f4;
}

.current-user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #f09433, #e6683c);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
}

.input-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.input-wrapper textarea {
  width: 100%;
  border: none;
  outline: none;
  font-size: 18px;
  resize: none;
  background: transparent;
}

.input-wrapper textarea::placeholder {
  color: #536471;
}

.input-wrapper .btn {
  align-self: flex-end;
  border-radius: 9999px;
  padding: 8px 20px;
  font-weight: 700;
}

.comments-list {
  padding: 0;
}

.empty-comments {
  padding: 40px 16px;
  text-align: center;
  color: #536471;
  font-size: 15px;
}

.comment-item {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid #eff3f4;
}

.comment-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 600;
  flex-shrink: 0;
  overflow: hidden;
  background-size: cover;
  background-position: center;
  cursor: pointer;
  transition: transform 0.2s;
}

.comment-avatar:hover {
  transform: scale(1.05);
}

.comment-content {
  flex: 1;
}

.comment-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 4px;
}

.comment-author {
  font-weight: 700;
  font-size: 14px;
  cursor: pointer;
  transition: color 0.2s;
}

.comment-author:hover {
  color: #0095f6;
  text-decoration: underline;
}

.comment-time {
  font-size: 13px;
  color: #536471;
}

.comment-text {
  font-size: 15px;
  line-height: 20px;
  color: var(--text-primary);
}

.comment-author {
  color: var(--text-primary);
}

.comment-time {
  color: var(--text-secondary);
}

.comment-item {
  cursor: context-menu;
}

.context-menu {
  position: fixed;
  z-index: 9999;
  background: var(--bg-primary);
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
  min-width: 160px;
  padding: 8px;
  border: 1px solid var(--border-color);
  animation: menuFadeIn 0.15s ease;
}

@keyframes menuFadeIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-5px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.context-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  font-size: 14px;
  font-weight: 500;
}

.context-menu-item:hover {
  background: var(--bg-hover);
}

.context-menu-item.delete {
  color: #ef4444;
}

.context-menu-item svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

:deep(.dark) .context-menu {
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
}

:deep(.dark) .comment-author:hover {
  color: #60a5fa;
}
</style>
