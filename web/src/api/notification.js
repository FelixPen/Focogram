import api from './index'

export const getNotifications = (params) => {
  return api.get('/auth/notifications', { params })
}

export const markNotificationsAsRead = () => {
  return api.post('/auth/notifications/read')
}

export const deleteNotification = (id) => {
  return api.delete(`/auth/notification/${id}`)
}

export const batchDeleteNotifications = (ids) => {
  return api.post('/auth/notifications/batch-delete', { ids })
}

export const createConversation = (data) => {
  return api.post('/auth/message/conversation', data)
}

export const sendPrivateMessage = (receiver_id, conv_id, data) => {
  return api.post(`/auth/message/send/${receiver_id}/${conv_id}`, data)
}

export const getConversationMessages = (conv_id) => {
  return api.get(`/auth/message/conversation/${conv_id}`)
}
