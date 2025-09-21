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
import { useThemeStore } from './stores/theme'
import { useSystemStore } from './stores/system'

const themeStore = useThemeStore()
const systemStore = useSystemStore()

const theme = computed(() => (themeStore.isDark ? darkTheme : null))

// 应用启动时加载系统信息
onMounted(() => {
  systemStore.fetchSystemInfo()
})
</script>
