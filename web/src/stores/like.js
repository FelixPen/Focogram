import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useUserStore } from './user'
import { getUserLiked } from '../api/like'

export const useLikeStore = defineStore('like', () => {
  const userStore = useUserStore()

  const likeUserSet = ref(new Set())

  const fetchLikedPosts = async () => {
    if (!userStore.isLoggedIn) return
    try {
      const res = await getUserLiked(userStore.user.userid)
      const ids = (res.posts || []).map(p => p.postid)
      likeUserSet.value = new Set(ids)
    } catch (err) {
      console.error('拉取点赞列表失败:', err)
    }
  }

  const isLiked = (postid) => {
    return likeUserSet.value.has(postid)
  }

  const addLike = (postid) => {
    likeUserSet.value.add(postid)
  }

  const removeLike = (postid) => {
    likeUserSet.value.delete(postid)
  }

  const clearAll = () => {
    likeUserSet.value.clear()
  }

  return {
    likeUserSet,
    fetchLikedPosts,
    isLiked,
    addLike,
    removeLike,
    clearAll
  }
})
