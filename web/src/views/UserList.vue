<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { getUserInfo } from '../api/auth'
import { getMyFollowing, getMyFollowers, getUserFollowing, getUserFollowers, followUser, unfollowUser, checkFollow } from '../api/follow'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const listType = computed(() => route.params.type || 'following')
const targetUserId = computed(() => route.params.userid || userStore.user?.userid)
const isOwnProfile = computed(() => targetUserId.value === userStore.user?.userid)

const users = ref([])
const loading = ref(false)
const profileUser = ref(null)

const pageTitle = computed(() => {
  const isOwn = isOwnProfile.value
  if (listType.value === 'following') {
    return isOwn ? '我的关注' : 'TA的关注'
  }
  return isOwn ? '我的粉丝' : 'TA的粉丝'
})

const fetchUserList = async () => {
  loading.value = true
  try {
    let res
    if (listType.value === 'following') {
      if (isOwnProfile.value) {
        res = await getMyFollowing()
      } else {
        res = await getUserFollowing(targetUserId.value)
      }
    } else {
      if (isOwnProfile.value) {
        res = await getMyFollowers()
      } else {
        res = await getUserFollowers(targetUserId.value)
      }
    }
    
    const userIds = listType.value === 'following' ? res.following : res.followers
    const userList = []
    
    for (const uid of userIds) {
      try {
        const userRes = await getUserInfo({ userid: uid })
        const user = userRes.user
        
        if (targetUserId.value === userStore.user?.userid) {
          try {
            const followRes = await checkFollow(uid)
            user.isFollowing = followRes.isFollowing
          } catch {
            user.isFollowing = listType.value === 'following'
          }
        } else {
          user.isFollowing = false
        }
        
        userList.push(user)
      } catch (e) {
        console.error('获取用户信息失败:', e)
      }
    }
    
    users.value = userList
  } catch (err) {
    console.error('获取列表失败:', err)
  } finally {
    loading.value = false
  }
}

const goToProfile = (userid) => {
  router.push(`/profile/${userid}`)
}

const handleFollow = async (user, index) => {
  if (user.userid === userStore.user?.userid) return
  
  try {
    if (user.isFollowing) {
      await unfollowUser(user.userid)
      users.value[index].isFollowing = false
    } else {
      await followUser(user.userid)
      users.value[index].isFollowing = true
    }
  } catch (err) {
    console.error('操作失败:', err)
  }
}

onMounted(() => {
  if (!isOwnProfile.value) {
    router.push(`/profile/${targetUserId.value}`)
    return
  }
  fetchUserList()
})
</script>

<template>
  <div class="userlist-page">
    <div class="page-header">
      <button class="back-btn" @click="router.back()">←</button>
      <h1>{{ pageTitle }}</h1>
    </div>

    <div class="userlist-content">
      <div v-if="loading" class="loading-container">
        <div class="spinner"></div>
        <p style="color: #8e8e8e; margin-top: 12px;">加载中...</p>
      </div>

      <div v-else-if="users.length === 0" class="empty-state">
        <p style="font-size: 48px; margin-bottom: 16px;">👥</p>
        <h2 style="margin-bottom: 8px;">暂无用户</h2>
        <p style="color: #8e8e8e;">去发现更多有趣的人吧</p>
      </div>

      <div 
        v-for="(user, index) in users" 
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
.userlist-page {
  max-width: 600px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 20px 20px;
  border-bottom: 1px solid #eff3f4;
  margin-bottom: 20px;
}

.back-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: #f7f9fa;
  font-size: 18px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.back-btn:hover {
  background: #eff3f4;
}

.page-header h1 {
  font-size: 20px;
  font-weight: 700;
}

.userlist-content {
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
