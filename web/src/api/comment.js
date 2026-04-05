import api from './index'

export const createComment = (postid, data) => {
  return api.post(`/auth/comment/${postid}`, data)
}

export const deleteComment = (commentid) => {
  return api.delete(`/auth/comment/${commentid}`)
}

export const getPostComments = (postid) => {
  return api.get(`/comment/${postid}`)
}
