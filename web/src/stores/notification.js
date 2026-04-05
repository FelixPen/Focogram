import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useUserStore } from './user'
import { getNotifications, markNotificationsAsRead } from '../api/notification'

export const useNotificationStore = defineStore('notification', () => {
  const userStore = useUserStore()

  const ws = ref(null)
  const notifications = ref([])
  const messages = ref([])
  const unreadCount = ref(0)
  const unreadMessageCount = ref(0)
  const conversationUnreadCounts = ref({}) // { conversation_id: count }

  const processUnreadSummary = (data) => {
    unreadMessageCount.value = data.total_unread || 0
    if (data.unread_details && Array.isArray(data.unread_details)) {
      data.unread_details.forEach(detail => {
        conversationUnreadCounts.value[detail.conversation_id] = detail.count
      })
    }
  }

  const addNotification = (data) => {
    if (data.content_type === 'message_unread_summary') {
      processUnreadSummary(data)
      return
    }

    if (data.content_type === 'private_message') {
      unreadMessageCount.value++
      const convId = data.conversation_id
      if (conversationUnreadCounts.value[convId]) {
        conversationUnreadCounts.value[convId]++
      } else {
        conversationUnreadCounts.value[convId] = 1
      }
      window.dispatchEvent(new CustomEvent('newPrivateMessage', { detail: data }))
      return
    }

    if (data.content_type === 'message') {
      const exists = messages.value.some(m => m.id === data.id)
      if (!exists) {
        messages.value.unshift(data)
        if (!data.is_read) {
          unreadMessageCount.value++
        }
      }
      return
    }

    const exists = notifications.value.some(n => n.id === data.id)
    if (!exists) {
      notifications.value.unshift(data)
      if (!data.is_read) {
        unreadCount.value++
      }
    }
  }

  const clearConversationUnread = (convId) => {
    const count = conversationUnreadCounts.value[convId] || 0
    if (count > 0) {
      unreadMessageCount.value = Math.max(0, unreadMessageCount.value - count)
      conversationUnreadCounts.value[convId] = 0
    }
  }

  const fetchOfflineNotifications = async () => {
    try {
      const res = await getNotifications()
      if (res.notifications) {
        for (let i = res.notifications.length - 1; i >= 0; i--) {
          addNotification(res.notifications[i])
        }
      }
    } catch (err) {
      console.error('拉取离线通知失败:', err)
    }
  }

  const markAllAsRead = async () => {
    try {
      await markNotificationsAsRead()
      unreadCount.value = 0
      notifications.value.forEach(n => {
        n.is_read = true
      })
    } catch (err) {
      console.error('标记已读失败:', err)
    }
  }

  const connectWebSocket = () => {
    if (!userStore.isLoggedIn) return
    if (ws.value) return

    const token = userStore.token
    const wsUrl = `ws://localhost:8080/ws?token=${token}`
    
    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      console.log('✅ WebSocket 实时通知已连接')
      fetchOfflineNotifications()
    }

    ws.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        addNotification(data)
      } catch (e) {
        console.error('❌ 解析WebSocket消息失败:', e)
      }
    }

    ws.value.onerror = (error) => {
      console.error('❌ WebSocket 连接错误:', error)
    }

    ws.value.onclose = (event) => {
      console.log('🔌 WebSocket 断开，3秒后重连...')
      ws.value = null
      setTimeout(() => {
        if (userStore.isLoggedIn) {
          connectWebSocket()
        }
      }, 3000)
    }
  }

  const disconnectWebSocket = () => {
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
  }

  const clearAll = () => {
    notifications.value = []
    unreadCount.value = 0
  }

  const markMessagesAsRead = () => {
    unreadMessageCount.value = 0
    messages.value.forEach(m => {
      m.is_read = true
    })
  }

  return {
    ws,
    notifications,
    messages,
    unreadCount,
    unreadMessageCount,
    conversationUnreadCounts,
    connectWebSocket,
    disconnectWebSocket,
    markAsRead: markAllAsRead,
    markMessagesAsRead,
    clearConversationUnread,
    clearAll
  }
})
