import api from './index'

export const likePost = (postid) => {
  return api.post(`/auth/like/${postid}`)
}

export const unlikePost = (postid) => {
  return api.post(`/auth/unlike/${postid}`)
}

export const getPostLikeCount = (postid) => {
  return api.get(`/like/count/${postid}`)
}

export const getPostLikeUsers = (postid) => {
  return api.get(`/auth/like/${postid}`)
}

export const getUserLiked = (userid) => {
  return api.get(`/post/liked/${userid}`)
}
