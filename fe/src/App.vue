<template>
  <n-config-provider :theme="theme">
    <n-global-style />
    <n-message-provider>
      <n-notification-provider>
        <n-dialog-provider>
          <router-view />
        </n-dialog-provider>
      </n-notification-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { darkTheme } from 'naive-ui'
import { useThemeStore, useSystemStore } from '@/stores'
import router from './router'

const themeStore = useThemeStore()
const systemStore = useSystemStore()

const theme = computed(() => (themeStore.isDark ? darkTheme : null))

// 应用启动时初始化主题和启动系统信息自动刷新
onMounted(() => {
  themeStore.initTheme()
  systemStore.load()
  systemStore.refresh().then((res) => {
    if (!res?.data?.initialized) {
      router.replace('/@init')
    }
  })
})
</script>
