import axios from 'axios'
import { localStg } from './storage'

// 响应数据类型
export interface ApiResponse<T = unknown> {
  msg: string
  code: number
  data?: T
}

// Auth Store 类型定义
interface AuthStore {
  accessToken: string
  tryRefreshToken: () => Promise<Models.UserInfo | null>
  userLogout: () => void
}

// 创建 axios 实例
export const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 用于存储正在刷新token的Promise，避免重复刷新
let refreshTokenPromise: Promise<string | null> | null = null

// 获取auth store的函数（延迟导入避免循环依赖）
let getAuthStore: (() => AuthStore) | null = null

// 设置auth store获取函数
export function setAuthStoreGetter(getter: () => AuthStore) {
  getAuthStore = getter
}

// 刷新token的函数
async function refreshAccessToken(): Promise<string | null> {
  if (!getAuthStore) {
    console.error('Auth store getter not set')
    return null
  }

  const authStore = getAuthStore()

  // 如果已经有正在进行的刷新请求，直接返回该Promise
  if (refreshTokenPromise) {
    return refreshTokenPromise
  }

  refreshTokenPromise = (async () => {
    try {
      const user = await authStore.tryRefreshToken()
      if (user && authStore.accessToken) {
        return authStore.accessToken
      }
      return null
    } catch (error) {
      console.error('刷新token失败:', error)
      return null
    } finally {
      refreshTokenPromise = null
    }
  })()

  return refreshTokenPromise
}

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 从存储获取 token
    const token = localStg.get('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    const { data } = response

    // 检查 HTTP 状态码
    if (response.status < 200 || response.status >= 300) {
      console.error('请求失败:', data.msg || '请求失败')
      return Promise.reject(new Error(data.msg || '请求失败'))
    }

    return response
  },
  async (error) => {
    const originalRequest = error.config

    if (error.response) {
      const { status, data } = error.response

      switch (status) {
        case 401:
          // 如果是刷新token的请求失败，直接跳转登录页
          if (originalRequest.url?.includes('/auth/refresh')) {
            console.error('刷新token失败，请重新登录')
            if (getAuthStore) {
              getAuthStore().userLogout()
            }
            window.location.href = '/@login'
            break
          }

          // 避免重复刷新
          if (!originalRequest._retry) {
            originalRequest._retry = true

            try {
              const newToken = await refreshAccessToken()
              if (newToken) {
                // 更新请求头中的token
                originalRequest.headers.Authorization = `Bearer ${newToken}`
                // 重新发送原始请求
                return api(originalRequest)
              } else {
                // 刷新失败，跳转到登录页
                console.error('token刷新失败，请重新登录')
                if (getAuthStore) {
                  getAuthStore().userLogout()
                }
                window.location.href = '/@login'
              }
            } catch (refreshError) {
              console.error('刷新token过程中出错:', refreshError)
              if (getAuthStore) {
                getAuthStore().userLogout()
              }
              window.location.href = '/@login'
            }
          }
          break
        case 403:
          console.error('权限不足')
          break
        case 404:
          console.error('请求的资源不存在')
          break
        case 500:
          console.error('服务器内部错误')
          break
        default:
          console.error('请求失败:', data?.msg || '请求失败')
      }
    } else if (error.request) {
      console.error('网络错误，请检查网络连接')
    } else {
      console.error('请求配置错误')
    }

    return Promise.reject(error)
  }
)

export default api
