import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useModalStore = defineStore('modal', () => {
  const showAlert = ref(false)
  const alertMessage = ref('')
  const alertType = ref('success')

  const showConfirm = ref(false)
  const confirmMessage = ref('')
  const confirmResolve = ref(null)

  const openAlert = (message, type = 'success') => {
    alertMessage.value = message
    alertType.value = type
    showAlert.value = true
  }

  const closeAlert = () => {
    showAlert.value = false
  }

  const openConfirm = (message) => {
    return new Promise((resolve) => {
      confirmMessage.value = message
      confirmResolve.value = resolve
      showConfirm.value = true
    })
  }

  const handleConfirm = (result) => {
    showConfirm.value = false
    if (confirmResolve.value) {
      confirmResolve.value(result)
      confirmResolve.value = null
    }
  }

  return {
    showAlert,
    alertMessage,
    alertType,
    showConfirm,
    confirmMessage,
    openAlert,
    closeAlert,
    openConfirm,
    handleConfirm
  }
})
