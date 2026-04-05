import api from './index'

export const followUser = (userid) => {
  return api.post(`/auth/followuser/${userid}`)
}

export const unfollowUser = (userid) => {
  return api.post(`/auth/unfollowuser/${userid}`)
}

export const getMyFollowing = () => {
  return api.get('/auth/following')
}

export const getMyFollowers = () => {
  return api.get('/auth/followers')
}

export const getUserFollowing = (userid) => {
  return api.get(`/user/following/${userid}`)
}

export const getUserFollowers = (userid) => {
  return api.get(`/user/followers/${userid}`)
}

export const checkFollow = (userid) => {
  return api.get(`/auth/checkfollow/${userid}`)
}
