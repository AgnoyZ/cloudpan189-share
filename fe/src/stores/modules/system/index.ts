import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getSystemInfo, type SystemInfo } from '@/api/setting'

// 默认系统信息
const defaultSystemInfo: SystemInfo = {
  initialized: true,
  enableAuth: true,
  title: '云盘分享系统',
  baseURL: 'http://localhost:5173',
  runTime: 62,
  runTimeHuman: '1分2秒',
}

export const useSystemStore = defineStore('system', () => {
  // 系统信息状态
  const systemInfo = ref<SystemInfo>(defaultSystemInfo)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 定时器相关
  let refreshTimer: NodeJS.Timeout | null = null
  const isAutoRefreshEnabled = ref(false)

  // 计算属性
  const isInitialized = computed(() => systemInfo.value.initialized)
  const isAuthEnabled = computed(() => systemInfo.value.enableAuth)
  const systemTitle = computed(() => systemInfo.value.title)
  const baseURL = computed(() => systemInfo.value.baseURL)
  const runTime = computed(() => systemInfo.value.runTime)
  const runTimeHuman = computed(() => systemInfo.value.runTimeHuman)

  // 获取系统信息
  const fetchSystemInfo = () => {
    if (loading.value) return Promise.resolve()

    loading.value = true
    error.value = null

    return getSystemInfo()
      .then((response) => {
        if (response.data) {
          systemInfo.value = response.data
        } else {
          error.value = response.msg || '获取系统信息失败'
        }
      })
      .catch((err) => {
        error.value = err instanceof Error ? err.message : '网络错误'
        console.error('获取系统信息失败:', err)
      })
      .finally(() => {
        loading.value = false
      })
  }

  // 刷新系统信息
  const refreshSystemInfo = () => {
    return fetchSystemInfo()
  }

  // 清空系统信息
  const clearSystemInfo = () => {
    systemInfo.value = { ...defaultSystemInfo }
    error.value = null
  }

  // 更新系统信息（用于初始化后更新状态）
  const updateSystemInfo = (info: Partial<SystemInfo>) => {
    systemInfo.value = { ...systemInfo.value, ...info }
  }

  // 启动自动刷新
  const startAutoRefresh = () => {
    if (isAutoRefreshEnabled.value) return

    isAutoRefreshEnabled.value = true

    // 立即获取一次系统信息
    fetchSystemInfo()

    // 设置定时器，每30秒刷新一次
    refreshTimer = setInterval(() => {
      fetchSystemInfo()
    }, 30000)
  }

  // 停止自动刷新
  const stopAutoRefresh = () => {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
    isAutoRefreshEnabled.value = false
  }

  return {
    // 状态
    systemInfo,
    loading,
    error,
    isAutoRefreshEnabled,

    // 计算属性
    isInitialized,
    isAuthEnabled,
    systemTitle,
    baseURL,
    runTime,
    runTimeHuman,

    // 方法
    fetchSystemInfo,
    refreshSystemInfo,
    clearSystemInfo,
    updateSystemInfo,
    startAutoRefresh,
    stopAutoRefresh,
  }
})
