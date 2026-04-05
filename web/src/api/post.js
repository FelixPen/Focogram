import api from './index'

export const uploadPostImage = (file) => {
  const formData = new FormData()
  formData.append('image', file)
  return api.post('/auth/upload/image', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const createPost = (data) => {
  return api.post('/auth/post', data)
}

export const deletePost = (postid) => {
  return api.delete(`/auth/post/${postid}`)
}

export const getUserPosts = (userid) => {
  return api.get(`/post/${userid}`)
}

export const getFollowingPosts = () => {
  return api.get('/auth/timeline')
}

export const getUserLikedPosts = (userid) => {
  return api.get(`/post/liked/${userid}`)
}

export const getPostDetail = (postid) => {
  return api.get(`/post/detail/${postid}`)
}
