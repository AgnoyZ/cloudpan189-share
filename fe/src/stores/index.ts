import { createPinia } from 'pinia'
import { setAuthStoreGetter } from '@/utils/api'

// 创建 pinia 实例
export const pinia = createPinia()

// 导出所有 stores
export { useAuthStore } from './modules/auth'
export { useSystemStore } from './modules/system'
export { useThemeStore } from './modules/theme'

// 初始化API和Auth Store的连接
export async function setupStores() {
  const { useAuthStore } = await import('./modules/auth')
  setAuthStoreGetter(() => useAuthStore())
}

// 默认导出 pinia 实例
export default pinia
