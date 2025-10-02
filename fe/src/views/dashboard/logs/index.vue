<template>
  <div class="logs-tabs-page">
    <n-card>
      <n-tabs v-model:value="active" type="line" animated>
        <n-tab-pane name="engine" tab="执行日志">
          <keep-alive>
            <EngineLogs v-if="active === 'engine'" />
          </keep-alive>
        </n-tab-pane>
        <n-tab-pane name="task" tab="任务日志">
          <keep-alive>
            <TaskLogs v-if="active === 'task'" />
          </keep-alive>
        </n-tab-pane>
        <n-tab-pane name="login" tab="登录日志">
          <keep-alive>
            <LoginLogs v-if="active === 'login'" />
          </keep-alive>
        </n-tab-pane>
      </n-tabs>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NTabs, NTabPane, NCard } from 'naive-ui'
import EngineLogs from '@/components/logs/EngineLogs.vue'
import TaskLogs from '@/components/logs/TaskLogs.vue'
import LoginLogs from '@/components/logs/LoginLogs.vue'

// 默认 激活 执行日志
const active = ref<'engine' | 'task' | 'login'>('engine')

// 若未来希望根据 query 或 hash 指定默认标签，可在此处理
onMounted(() => {
  // const url = new URL(window.location.href)
  // const tab = (url.searchParams.get('tab') || '').toLowerCase()
  // if (tab === 'task' || tab === 'login') active.value = tab
})
</script>

<style scoped>
.logs-tabs-page {
  padding: 0;
}
</style>
