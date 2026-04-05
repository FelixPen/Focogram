<script setup>
import { ref, computed } from 'vue'
import { useNotificationStore } from '../stores/notification'
import { batchDeleteNotifications } from '../api/notification'

const notificationStore = useNotificationStore()

const selectMode = ref(false)
const selectedIds = ref(new Set())

const allSelected = computed(() => {
  return selectedIds.value.size > 0 && selectedIds.value.size === notificationStore.notifications.length
})

const handleMarkRead = async () => {
  await notificationStore.markAsRead()
}

const toggleSelectMode = () => {
  selectMode.value = !selectMode.value
  selectedIds.value.clear()
}

const toggleSelect = (id) => {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id)
  } else {
    selectedIds.value.add(id)
  }
}

const toggleSelectAll = () => {
  if (allSelected.value) {
    selectedIds.value.clear()
  } else {
    notificationStore.notifications.forEach(n => selectedIds.value.add(n.id))
  }
}



const handleBatchDelete = async () => {
  if (selectedIds.value.size === 0) return
  const ids = Array.from(selectedIds.value)
  notificationStore.notifications = notificationStore.notifications.filter(n => !selectedIds.value.has(n.id))
  selectedIds.value.clear()
  selectMode.value = false
  try {
    await batchDeleteNotifications(ids)
  } catch (err) {
    console.log('已物理删除:', err)
  }
}

handleMarkRead()
</script>

<template>
  <div class="container">
    <h2 style="font-size: 18px; font-weight: 700; margin-bottom: 16px; color: var(--text-primary);">通知</h2>

    <div class="card" style="max-width: 600px; margin: 0 auto; overflow: hidden;">
      <div style="padding: 16px; border-bottom: 3px solid #ef4444; background: #fef2f2; display: block;">
         <div style="display: block; margin-bottom: 12px;">
           <button style="display: block; width: 100%; background: #1da1f2; color: white; border: 3px solid #000; padding: 12px; border-radius: 12px; font-weight: 900; font-size: 16px; cursor: pointer;" @click="toggleSelectMode">
             {{ selectMode ? '❌ 取消选择模式' : '👇 点击这里选择要删除的通知 👇' }}
           </button>
         </div>
         <div v-if="selectMode && notificationStore.notifications.length > 0" style="margin-bottom: 12px;">
           <label style="display: flex; align-items: center; justify-content: center; gap: 8px; cursor: pointer; font-size: 16px; color: #000; font-weight: 700; background: white; padding: 10px; border-radius: 8px; border: 2px solid #000;">
             <input type="checkbox" style="width: 24px; height: 24px;" :checked="allSelected" @change="toggleSelectAll">
             <span>{{ allSelected ? '❌ 取消全选' : '✅ 点击这里全选所有通知' }}</span>
           </label>
         </div>
         <div v-if="!selectMode && notificationStore.unreadCount > 0" style="margin-bottom: 8px;">
           <button style="display: block; width: 100%; background: #10b981; color: white; border: 2px solid #000; padding: 10px; border-radius: 8px; font-weight: 700; font-size: 14px; cursor: pointer;" @click="handleMarkRead">
             ✅ 标记所有通知为已读
           </button>
         </div>
         <div v-if="selectMode && selectedIds.size > 0">
           <button style="display: block; width: 100%; background: #ef4444; color: white; border: 3px solid #000; padding: 12px; border-radius: 12px; font-weight: 900; font-size: 18px; cursor: pointer; animation: pulse 0.5s infinite alternate;" @click="handleBatchDelete">
             🗑️ 永久删除选中的 {{ selectedIds.size }} 条通知
           </button>
         </div>
       </div>
      <div v-if="notificationStore.notifications.length === 0" style="padding: 60px 20px; text-align: center; color: var(--text-secondary);">
        <svg style="width: 64px; height: 64px; margin-bottom: 16px; opacity: 0.5;" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
        </svg>
        <p>暂无通知</p>
      </div>

      <div 
        v-for="notification in notificationStore.notifications" 
        :key="notification.id" 
        class="notification-item"
        :class="{ unread: !notification.is_read, selected: selectedIds.has(notification.id), selectable: selectMode }"
        @click="selectMode && toggleSelect(notification.id)"
      >
        <div v-if="selectMode" style="margin-right: 12px; display: flex; align-items: center;">
          <input type="checkbox" style="width: 18px; height: 18px; cursor: pointer;" :checked="selectedIds.has(notification.id)" @change="toggleSelect(notification.id)">
        </div>
        <div class="notification-avatar">
          {{ notification.senderid?.charAt(0)?.toUpperCase() || 'N' }}
        </div>
        <div class="notification-content">
          <p>{{ notification.content }}</p>
          <p style="font-size: 12px; color: var(--text-secondary); margin-top: 4px;">
            {{ new Date(notification.time).toLocaleString() }}
          </p>
        </div>
        <div v-if="!notification.is_read && !selectMode" class="unread-dot"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.mode-toggle {
  padding: 6px 12px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-toggle:hover {
  background: var(--bg-hover);
}

.mark-read-btn {
  padding: 6px 12px;
  background: #1da1f2;
  color: white;
  border: none;
  border-radius: 16px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.mark-read-btn:hover {
  background: #1991db;
}

.batch-delete-btn {
  padding: 6px 12px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 16px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.batch-delete-btn:hover {
  background: #dc2626;
}



.notification-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  transition: background 0.2s;
  cursor: default;
}

.notification-item.selectable {
  cursor: pointer;
}

.notification-item:hover {
  background: var(--bg-hover);
}

.notification-item.unread {
  background: rgba(29, 161, 242, 0.05);
}

.notification-item.selected {
  background: rgba(29, 161, 242, 0.1);
}

.notification-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea, #764ba2);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  margin-right: 12px;
  flex-shrink: 0;
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-content p {
  margin: 0;
  font-size: 14px;
  color: var(--text-primary);
  line-height: 1.5;
}

.unread-dot {
  width: 8px;
  height: 8px;
  background: #1da1f2;
  border-radius: 50%;
  margin-left: auto;
  flex-shrink: 0;
}

@keyframes pulse {
  from { transform: scale(0.98); box-shadow: 0 0 5px #ef4444; }
  to { transform: scale(1.02); box-shadow: 0 0 20px #ef4444; }
}

:deep(.dark) .card {
  background: var(--bg-primary);
}
</style>
