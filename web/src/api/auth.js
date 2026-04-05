import api from './index'

export const register = (data) => {
  return api.post('/register', data)
}

export const login = (data) => {
  return api.post('/login', data)
}

export const getUserInfo = (params) => {
  return api.get('/userinfo', { params })
}

export const updateUserInfo = (data) => {
  return api.patch('/auth/userinfo', data)
}

export const updatePassword = (data) => {
  return api.patch('/auth/password', data)
}

export const resetPasswordByEmail = (data) => {
  return api.post('/password/reset', data)
}

export const uploadAvatar = (file) => {
  const formData = new FormData()
  formData.append('avatar', file)
  return api.post('/auth/upload/avatar', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}
