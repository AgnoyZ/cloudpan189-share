import axios from 'axios'

// 创建 axios 实例
export const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 响应数据类型
export interface ApiResponse<T = unknown> {
  msg: string
  code: number
  data?: T
}

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token')
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
  (error) => {
    if (error.response) {
      const { status, data } = error.response

      switch (status) {
        case 401:
          console.error('未授权访问，请重新登录')
          // 清除本地存储的 token
          localStorage.removeItem('token')
          // 跳转到登录页
          window.location.href = '/@login'
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
