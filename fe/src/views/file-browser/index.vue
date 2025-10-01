<template>
  <div class="file-browser">
    <!-- 顶部导航栏 -->
    <div class="header">
      <div class="breadcrumb-section">
        <n-breadcrumb>
          <n-breadcrumb-item @click="navigateToPath('/')">
            <n-icon :component="HomeOutline" />
          </n-breadcrumb-item>
          <n-breadcrumb-item
            v-for="(item, index) in breadcrumbs"
            :key="index"
            @click="navigateToPath(item.href.replace('/api/file/open', ''))"
            :clickable="index < breadcrumbs.length - 1"
          >
            {{ item.name }}
          </n-breadcrumb-item>
        </n-breadcrumb>
      </div>

      <div class="actions">
        <n-button text @click="goBack" :disabled="!canGoBack">
          <template #icon>
            <n-icon :component="ArrowUndoOutline" />
          </template>
          返回上级
        </n-button>
        <n-button text @click="goAdmin">
          <template #icon>
            <n-icon :component="SettingsOutline" />
          </template>
          后台
        </n-button>
        <n-button text @click="showSearch = true">
          <template #icon>
            <n-icon :component="SearchOutline" />
          </template>
          搜索
        </n-button>
        <n-button text @click="refreshCurrentPath" :loading="loading">
          <template #icon>
            <n-icon :component="RefreshOutline" />
          </template>
          刷新
        </n-button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <n-spin size="large" />
    </div>

    <!-- 根据isDir判断渲染不同组件 -->
    <template v-else-if="fileInfo">
      <!-- 目录：显示文件列表 -->
      <div v-if="fileInfo.isDir" class="directory-view">
        <!-- 文件列表组件 -->
        <FileList
          :file-list="fileInfo.children || []"
          :loading="false"
          @file-click="handleFileClick"
          @download="downloadFile"
        />
      </div>
      <!-- 文件：显示文件详情 -->
      <FileDetail v-else :file-info="fileInfo" :loading="false" />
    </template>

    <!-- 搜索弹窗 -->
    <SearchFilesDialog
      :show="showSearch"
      :currentDirId="fileInfo && fileInfo.isDir ? fileInfo.id : null"
      @update:show="(v) => (showSearch = v)"
      @select="onSearchSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NIcon, NBreadcrumb, NBreadcrumbItem, NSpin, useMessage } from 'naive-ui'
import {
  HomeOutline,
  RefreshOutline,
  ArrowUndoOutline,
  SearchOutline,
  SettingsOutline,
} from '@vicons/ionicons5'
import {
  openFile,
  createDownloadUrl,
  type FileChild,
  type BreadcrumbItem,
  type FileOpenResponse,
  type FileSearchItem,
} from '@/api/file'
import { FileList, FileDetail, SearchFilesDialog } from '@/components/file-browser'

const route = useRoute()
const router = useRouter()
const message = useMessage()

// 响应式数据
const loading = ref(false)
const fileInfo = ref<FileOpenResponse | null>(null)
const currentPath = ref('/')
const breadcrumbs = ref<BreadcrumbItem[]>([])
const showSearch = ref(false)

// 计算属性
const canGoBack = computed(() => breadcrumbs.value.length > 0)

// 方法
const loadPath = (path: string) => {
  loading.value = true
  openFile(path)
    .then((response) => {
      if (response.code === 200 && response.data) {
        fileInfo.value = response.data
        currentPath.value = path
        breadcrumbs.value = response.data.breadcrumbs || []
      } else {
        message.error(response.msg || '加载失败')
      }
    })
    .catch((error) => {
      console.error('加载文件失败:', error)
      message.error('加载文件失败')
    })
    .finally(() => {
      loading.value = false
    })
}

const handleFileClick = (file: FileChild) => {
  navigateToPath(file.href)
}

const navigateToPath = (path: string) => {
  router.push({
    path: '/',
    query: { path: path },
  })
}

const refreshCurrentPath = () => {
  loadPath(currentPath.value)
}

const goAdmin = () => {
  // 后台首页路由，如有自定义请改成你的后台路径
  router.push({ path: '/@dashboard' })
}

const onSearchSelect = (row: FileSearchItem) => {
  // 目录：直接进入目录；文件：进入其所在目录
  let targetPath = row.fullPath || '/'
  if (!row.isDir) {
    const idx = targetPath.lastIndexOf('/')
    targetPath = idx > 0 ? targetPath.slice(0, idx) : '/'
  }
  showSearch.value = false
  navigateToPath(targetPath)
}

const goBack = () => {
  if (breadcrumbs.value.length > 0) {
    const parentIndex = breadcrumbs.value.length - 2
    if (parentIndex >= 0) {
      const parentPath = breadcrumbs.value[parentIndex].href.replace('/api/file/open', '')
      navigateToPath(parentPath)
    } else {
      navigateToPath('/')
    }
  }
}

const downloadFile = (file: FileChild) => {
  createDownloadUrl({ fileId: file.id })
    .then((response) => {
      if (response.code === 200 && response.data) {
        const link = document.createElement('a')
        link.href = response.data.downloadUrl
        link.download = file.name
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        message.success('开始下载')
      } else {
        message.error(response.msg || '创建下载链接失败')
      }
    })
    .catch((error) => {
      console.error('下载失败:', error)
      message.error('下载失败')
    })
}

// 监听路由变化
watch(
  () => route.query.path,
  (newPath) => {
    const path = (newPath as string) || '/'
    currentPath.value = path
    loadPath(path)
  },
  { immediate: true }
)
</script>

<style scoped>
.file-browser {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  background: var(--n-color-target);
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: var(--n-card-color);
  border-radius: 8px;
  margin-bottom: 20px;
  border: 1px solid var(--n-border-color);

  /* 使顶部栏在页面滚动时保持固定在顶部 */
  position: sticky;
  top: 0;
  z-index: 100;

  /* 贴顶时更清晰的层级与视觉分隔 */
  box-shadow: 0 2px 8px rgb(0 0 0 / 6%);
  backdrop-filter: saturate(180%) blur(2px);
}

.breadcrumb-section {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.actions {
  display: flex;
  gap: 8px;
}

.file-container {
  background: var(--n-card-color);
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  overflow: hidden;
}

.list-header {
  display: grid;
  grid-template-columns: 1fr 120px 180px 120px;
  gap: 16px;
  padding: 12px 20px;
  background: var(--n-color-hover);
  border-bottom: 1px solid var(--n-border-color);
  font-weight: 500;
  color: var(--n-text-color);
  font-size: 14px;
}

.header-item {
  display: flex;
  align-items: center;
}

.header-item:nth-child(2),
.header-item:nth-child(3),
.header-item:nth-child(4) {
  justify-content: center;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 60px 20px;
}

.file-list {
  min-height: 400px;
}

.file-item {
  border-bottom: 1px solid var(--n-divider-color);
  cursor: pointer;
  transition: background-color 0.2s;
}

.file-item:hover {
  background-color: var(--n-color-hover);
}

.file-item:last-child {
  border-bottom: none;
}

.file-info {
  display: grid;
  grid-template-columns: 1fr 120px 180px 120px;
  gap: 16px;
  padding: 12px 20px;
  align-items: center;
}

.file-icon-name {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.file-icon {
  font-size: 20px;
  color: var(--n-primary-color);
  flex-shrink: 0;
}

.file-name {
  font-size: 14px;
  color: var(--n-text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-size {
  font-size: 14px;
  color: var(--n-text-color-2);
  text-align: center;
}

.file-date {
  font-size: 14px;
  color: var(--n-text-color-2);
  text-align: center;
}

.file-actions {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.empty-state {
  padding: 60px 20px;
  text-align: center;
}

.file-preview {
  padding: 16px 0;
}

.preview-actions {
  margin-top: 24px;
}

/* 响应式设计 */
@media (width <= 768px) {
  .file-browser {
    padding: 12px;
  }

  .header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .actions {
    justify-content: center;
  }

  .list-header,
  .file-info {
    grid-template-columns: 1fr 80px 100px;
    gap: 8px;
    padding: 8px 12px;
  }

  .file-actions {
    flex-direction: column;
    gap: 4px;
  }

  .file-date {
    display: none;
  }
}

@media (width <= 480px) {
  .list-header,
  .file-info {
    grid-template-columns: 1fr 60px;
    gap: 8px;
  }

  .file-size {
    display: none;
  }

  .file-name {
    font-size: 13px;
  }
}
</style>
