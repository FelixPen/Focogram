import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login, register, getUserInfo } from '../api/auth'

const safeParseJSON = (str, defaultValue = null) => {
  try {
    if (!str || str === 'undefined' || str === 'null') {
      return defaultValue
    }
    return JSON.parse(str)
  } catch (e) {
    return defaultValue
  }
}

if (localStorage.getItem('user') === 'undefined') {
  localStorage.removeItem('user')
  localStorage.removeItem('token')
}

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(safeParseJSON(localStorage.getItem('user')))

  const isLoggedIn = computed(() => !!token.value && !!user.value)

  const setToken = (newToken) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUser = (newUser) => {
    user.value = newUser
    localStorage.setItem('user', JSON.stringify(newUser))
  }

  const logout = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const handleLogin = async (credentials) => {
    const res = await login(credentials)
    setToken(res.token)
    setUser(res.user)
    return res
  }

  const handleRegister = async (data) => {
    const res = await register(data)
    return res
  }

  const fetchUserInfo = async (userid) => {
    const res = await getUserInfo({ userid })
    return res
  }

  return {
    token,
    user,
    isLoggedIn,
    setToken,
    setUser,
    logout,
    handleLogin,
    handleRegister,
    fetchUserInfo
  }
})
