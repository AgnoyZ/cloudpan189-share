import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getUserInfo } from '@/api/user'
import {
  login,
  refreshToken,
  type LoginRequest,
  type LoginResponse,
  type RefreshRequest,
} from '@/api/auth'
import type { UserInfo } from '@/types/models'
import type { ApiResponse } from '@/utils/api'

export const useUserStore = defineStore('user', () => {
  // 状态
  const user = ref<UserInfo | null>(null)
  const accessToken = ref<string>('')
  const refreshTokenValue = ref<string>('')
  const isLoggedIn = ref<boolean>(false)
  const loading = ref<boolean>(false)

  // 计算属性
  const isAdmin = computed(() => user.value?.isAdmin || false)
  const username = computed(() => user.value?.username || '')
  const userId = computed(() => user.value?.id || 0)

  // 初始化用户状态（从localStorage恢复）
  const initUserState = (): Promise<void> => {
    const token = localStorage.getItem('token')
    const refreshTokenLocal = localStorage.getItem('refreshToken')
    const userLocal = localStorage.getItem('user')

    if (token && refreshTokenLocal && userLocal) {
      accessToken.value = token
      refreshTokenValue.value = refreshTokenLocal

      // 解析用户信息
      const parseUser = () => {
        return JSON.parse(userLocal)
      }

      return Promise.resolve()
        .then(() => parseUser())
        .then((parsedUser) => {
          user.value = parsedUser
          isLoggedIn.value = true
        })
        .catch((error) => {
          console.error('解析用户信息失败:', error)
          clearUserState()
        })
    }

    return Promise.resolve()
  }

  // 清除用户状态
  const clearUserState = () => {
    user.value = null
    accessToken.value = ''
    refreshTokenValue.value = ''
    isLoggedIn.value = false
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('user')
  }

  // 保存用户状态到localStorage
  const saveUserState = (loginResponse: LoginResponse) => {
    accessToken.value = loginResponse.accessToken
    refreshTokenValue.value = loginResponse.refreshToken
    user.value = loginResponse.user
    isLoggedIn.value = true

    localStorage.setItem('token', loginResponse.accessToken)
    localStorage.setItem('refreshToken', loginResponse.refreshToken)
    localStorage.setItem('user', JSON.stringify(loginResponse.user))
  }

  // 用户登录
  const userLogin = (loginData: LoginRequest): Promise<ApiResponse<LoginResponse>> => {
    loading.value = true

    return login(loginData)
      .then((response) => {
        if (response.data) {
          saveUserState(response.data)
        }
        return response
      })
      .finally(() => {
        loading.value = false
      })
  }

  // 用户登出
  const userLogout = () => {
    clearUserState()
  }

  // 获取用户信息
  const fetchUserInfo = (): Promise<UserInfo | null> => {
    if (!isLoggedIn.value) {
      return Promise.resolve(null)
    }

    loading.value = true

    return getUserInfo()
      .then((response) => {
        if (response.data) {
          user.value = response.data
          localStorage.setItem('user', JSON.stringify(response.data))
          return response.data
        }
        return null
      })
      .catch((error) => {
        console.error('获取用户信息失败:', error)
        // 如果获取用户信息失败，可能是token过期，尝试刷新token
        if (refreshTokenValue.value) {
          return tryRefreshToken()
        }
        return null
      })
      .finally(() => {
        loading.value = false
      })
  }

  // 刷新访问令牌
  const tryRefreshToken = (): Promise<UserInfo | null> => {
    if (!refreshTokenValue.value) {
      clearUserState()
      return Promise.resolve(null)
    }

    const refreshData: RefreshRequest = {
      refreshToken: refreshTokenValue.value,
    }

    return refreshToken(refreshData)
      .then((response) => {
        if (response.data) {
          saveUserState(response.data)
          return response.data.user
        }
        return null
      })
      .catch((error) => {
        console.error('刷新token失败:', error)
        clearUserState()
        return null
      })
  }

  // 更新用户信息（本地更新，不调用API）
  const updateUserInfo = (newUserInfo: Partial<UserInfo>) => {
    if (user.value) {
      user.value = { ...user.value, ...newUserInfo }
      localStorage.setItem('user', JSON.stringify(user.value))
    }
  }

  // 检查登录状态
  const checkLoginStatus = (): Promise<boolean> => {
    if (!isLoggedIn.value || !accessToken.value) {
      return Promise.resolve(false)
    }

    // 如果已经有用户信息，直接返回true，避免重复请求
    if (user.value) {
      return Promise.resolve(true)
    }

    // 只有在没有用户信息时才获取用户信息来验证token是否有效
    return fetchUserInfo().then((userInfo) => {
      return userInfo !== null
    })
  }

  return {
    // 状态
    user,
    accessToken,
    refreshTokenValue,
    isLoggedIn,
    loading,

    // 计算属性
    isAdmin,
    username,
    userId,

    // 方法
    initUserState,
    clearUserState,
    userLogin,
    userLogout,
    fetchUserInfo,
    tryRefreshToken,
    updateUserInfo,
    checkLoginStatus,
  }
})
