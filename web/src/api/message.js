import api from './index'

export const createConversation = (data) => {
  return api.post('/auth/message/conversation', {
    target_user_id: data.receiver_id
  })
}

export const sendMessage = (receiver_id, conv_id, content) => {
  return api.post(`/auth/message/send/${receiver_id}/${conv_id}`, { content })
}

export const getConversationMessages = (conv_id) => {
  return api.get(`/auth/message/conversation/${conv_id}`)
}

export const getConversations = () => {
  return api.get('/auth/message/conversations')
}

export const markConversationAsRead = (conv_id) => {
  return api.post(`/auth/message/conversation/${conv_id}/read`)
}

export const getUnreadMessageStats = () => {
  return api.get('/auth/message/unread/stats')
}
