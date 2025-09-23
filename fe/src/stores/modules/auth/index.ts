import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { getUserInfo } from '@/api/user'
import {
  login,
  refreshToken,
  type LoginRequest,
  type LoginResponse,
  type RefreshRequest,
} from '@/api/auth'
import type { ApiResponse } from '@/utils/api'
import { localStg } from '@/utils/storage'

// 使用全局类型定义
type UserInfo = Models.UserInfo

// Token刷新提前时间（5分钟）
const TOKEN_REFRESH_BUFFER = 5 * 60 * 1000
// Token自动刷新阈值（60分钟）
const TOKEN_AUTO_REFRESH_THRESHOLD = 60 * 60 * 1000

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const user = ref<UserInfo | null>(null)
  const accessToken = ref<string>('')
  const refreshTokenValue = ref<string>('')
  const loading = ref<boolean>(false)
  const expireTime = ref<number>(0)

  // 计算属性
  const isLogin = computed(() => !!accessToken.value && expireTime.value > Date.now())
  const isAdmin = computed(() => user.value?.isAdmin || false)
  const username = computed(() => user.value?.username || '')
  const userId = computed(() => user.value?.id || 0)
  // 需要刷新token（距离过期时间小于60分钟）
  const requireRefreshToken = computed(
    () => expireTime.value - Date.now() < TOKEN_AUTO_REFRESH_THRESHOLD
  )

  // 清除用户状态
  function clearUserState() {
    user.value = null
    accessToken.value = ''
    refreshTokenValue.value = ''
    expireTime.value = 0

    localStg.remove('token')
    localStg.remove('refreshToken')
    localStg.remove('user')
    localStg.remove('expireTime')
  }

  // 保存用户状态到localStorage
  function saveUserState(loginResponse: LoginResponse) {
    const expirationTime = Date.now() + loginResponse.expiresIn * 1000 - TOKEN_REFRESH_BUFFER

    // 更新内存状态
    accessToken.value = loginResponse.accessToken
    refreshTokenValue.value = loginResponse.refreshToken
    user.value = loginResponse.user
    expireTime.value = expirationTime

    // 保存到本地存储
    localStg.set('token', loginResponse.accessToken)
    localStg.set('refreshToken', loginResponse.refreshToken)
    localStg.set('user', loginResponse.user)
    localStg.set('expireTime', expirationTime)
  }

  // 初始化用户状态（从localStorage恢复）
  async function initUserState(): Promise<UserInfo | null> {
    const token = localStg.get('token')
    const refreshTokenLocal = localStg.get('refreshToken')
    const userLocal = localStg.get('user')
    const expireTimeLocal = localStg.get('expireTime')

    if (token && refreshTokenLocal && userLocal && expireTimeLocal) {
      accessToken.value = token
      refreshTokenValue.value = refreshTokenLocal
      user.value = userLocal
      expireTime.value = expireTimeLocal
    }

    return fetchUserInfo()
  }

  // 用户登录
  async function userLogin(loginData: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    loading.value = true

    try {
      const response = await login(loginData)
      if (response.data) {
        saveUserState(response.data)
      }
      return response
    } finally {
      loading.value = false
    }
  }

  // 用户登出
  function userLogout() {
    clearUserState()
  }

  // 获取用户信息
  async function fetchUserInfo(): Promise<UserInfo | null> {
    loading.value = true

    try {
      const response = await getUserInfo()
      if (response.data) {
        user.value = response.data
        localStg.set('user', response.data)
        return response.data
      }
      return null
    } catch (error) {
      console.error('获取用户信息失败:', error)
      // 如果获取用户信息失败，可能是token过期，尝试刷新token
      if (refreshTokenValue.value) {
        return await tryRefreshToken()
      }
      return null
    } finally {
      loading.value = false
    }
  }

  // 刷新访问令牌
  async function tryRefreshToken(): Promise<UserInfo | null> {
    if (!refreshTokenValue.value) {
      clearUserState()
      return null
    }

    loading.value = true

    try {
      const refreshData: RefreshRequest = {
        refreshToken: refreshTokenValue.value,
      }

      const response = await refreshToken(refreshData)
      if (response.data) {
        saveUserState(response.data)
        return response.data.user
      }
      return null
    } catch (error) {
      console.error('刷新token失败:', error)
      clearUserState()
      return null
    } finally {
      loading.value = false
    }
  }

  // 更新用户信息（本地更新，不调用API）
  function updateUserInfo(newUserInfo: Partial<UserInfo>) {
    if (user.value) {
      user.value = { ...user.value, ...newUserInfo }
      localStg.set('user', user.value)
    }
  }

  return {
    // 状态
    user,
    accessToken,
    refreshTokenValue,
    isLogin,
    loading,
    requireRefreshToken,

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
  }
})
