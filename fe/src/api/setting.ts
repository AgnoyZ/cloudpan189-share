import { api, type ApiResponse } from '@/utils/api'

// 系统初始化请求接口
export interface InitSystemRequest {
  baseURL: string // 系统基础URL
  enableAuth: boolean // 是否启用认证
  superUsername: string // 超级管理员用户名，长度3-20位
  superPassword: string // 超级管理员密码，长度6-20位
  title: string // 系统标题
}

// ===== 系统设置接口 =====

// 获取系统信息
export const getSystemInfo = (): Promise<ApiResponse<Models.SystemInfo>> => {
  return api.get('/setting/info').then((res) => res.data)
}

// 初始化系统
export const initSystem = (data: InitSystemRequest): Promise<ApiResponse> => {
  return api.post('/setting/init_system', data).then((res) => res.data)
}
