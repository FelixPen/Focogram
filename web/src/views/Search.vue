<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { getUserInfo } from '../api/auth'
import { followUser, unfollowUser, checkFollow } from '../api/follow'

const router = useRouter()
const userStore = useUserStore()

const searchKeyword = ref('')
const searchResults = ref([])
const searching = ref(false)

const handleSearch = async () => {
  if (!searchKeyword.value.trim()) {
    searchResults.value = []
    return
  }
  
  searching.value = true
  try {
    const res = await getUserInfo({ keyword: searchKeyword.value })
    const users = res.users || []
    
    for (const user of users) {
      try {
        const followRes = await checkFollow(user.userid)
        user.isFollowing = followRes.isFollowing
      } catch {
        user.isFollowing = false
      }
    }
    
    searchResults.value = users
  } catch (err) {
    searchResults.value = []
  } finally {
    searching.value = false
  }
}

const goToProfile = (userid) => {
  console.log('跳转到用户主页:', userid)
  if (userid) {
    router.push(`/profile/${userid}`)
  } else {
    console.error('用户ID为空')
  }
}

const handleFollow = async (user, index) => {
  if (user.userid === userStore.user?.userid) return
  
  try {
    if (user.isFollowing) {
      await unfollowUser(user.userid)
      searchResults.value[index].isFollowing = false
    } else {
      await followUser(user.userid)
      searchResults.value[index].isFollowing = true
    }
  } catch (err) {
    console.error('操作失败:', err)
  }
}
</script>

<template>
  <div class="search-page">
    <div class="page-header">
      <h1>🔍 搜索用户</h1>
    </div>

    <div class="search-bar">
      <input 
        type="text" 
        v-model="searchKeyword"
        placeholder="输入用户名或账号搜索..."
        class="search-input"
        @keyup.enter="handleSearch"
      >
      <button v-if="!searching" class="search-btn" @click="handleSearch">🔍 搜索</button>
      <span v-else class="search-spinner-page"></span>
    </div>

    <div class="search-results">
      <div v-if="searching" class="loading-container">
        <div class="spinner"></div>
        <p style="color: #8e8e8e; margin-top: 12px;">搜索中...</p>
      </div>

      <div v-else-if="searchKeyword && searchResults.length === 0" class="empty-state">
        <p style="font-size: 48px; margin-bottom: 16px;">🔍</p>
        <h2 style="margin-bottom: 8px;">未找到用户</h2>
        <p style="color: #8e8e8e;">试试其他关键词吧</p>
      </div>

      <div v-else-if="!searchKeyword" class="empty-state">
        <p style="font-size: 48px; margin-bottom: 16px;">👋</p>
        <h2 style="margin-bottom: 8px;">搜索发现</h2>
        <p style="color: #8e8e8e;">输入关键词开始搜索用户</p>
      </div>

      <div 
        v-for="(user, index) in searchResults" 
        :key="user.userid"
        class="user-card card"
        @click="goToProfile(user.userid)"
      >
        <div class="user-info">
          <div v-if="user.avatarUrl" class="user-avatar" :style="{ backgroundImage: `url(${user.avatarUrl})`, backgroundSize: 'cover' }"></div>
          <div v-else class="user-avatar">{{ user.username?.charAt(0)?.toUpperCase() || 'U' }}</div>
          
          <div class="user-details">
            <h3 class="user-username">{{ user.username }}</h3>
            <p class="user-userid">ID: {{ user.userid }}</p>
            <p v-if="user.describe" class="user-bio">{{ user.describe }}</p>
          </div>
        </div>

        <button 
          v-if="user.userid !== userStore.user?.userid"
          class="follow-btn-small"
          :class="{ following: user.isFollowing }"
          @click.stop="handleFollow(user, index)"
        >
          {{ user.isFollowing ? '已关注' : '关注' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-page {
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

.search-bar {
  position: relative;
  padding: 0 20px 20px;
  display: flex;
  gap: 12px;
  align-items: center;
}

.search-input {
  flex: 1;
  padding: 14px 20px;
  border: 1px solid #dbdbdb;
  border-radius: 9999px;
  font-size: 16px;
  background: #f7f9fa;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #1d9bf0;
  background: white;
}

.search-btn {
  padding: 14px 28px;
  border-radius: 9999px;
  border: none;
  background: #1d9bf0;
  color: white;
  font-weight: 700;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.search-btn:hover {
  background: #1a8cd8;
}

.search-spinner-page {
  width: 24px;
  height: 24px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #1d9bf0;
  border-radius: 50%;
  animation: spin-center 0.8s linear infinite;
  margin-right: 10px;
}

@keyframes spin-center {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.search-results {
  padding: 0 20px;
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
  animation: spin-center 0.8s linear infinite;
}

@keyframes spin-center {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.user-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: background 0.2s;
}

.user-card:hover {
  background: #f7f9fa;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  cursor: pointer;
}

.user-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: #1d9bf0;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 24px;
  flex-shrink: 0;
}

.user-details {
  flex: 1;
  min-width: 0;
}

.user-username {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 4px;
}

.user-userid {
  font-size: 13px;
  color: #536471;
  margin-bottom: 4px;
}

.user-bio {
  font-size: 14px;
  color: #536471;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.follow-btn-small {
  padding: 8px 20px;
  border-radius: 9999px;
  border: none;
  background: #1d9bf0;
  color: white;
  font-weight: 700;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.follow-btn-small:hover {
  background: #1a8cd8;
}

.follow-btn-small.following {
  background: transparent;
  color: #0f1419;
  border: 1px solid #cfd9de;
}

.follow-btn-small.following:hover {
  border-color: #f4212e;
  color: #f4212e;
  background: rgba(244, 33, 46, 0.1);
}
</style>
