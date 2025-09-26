<template>
  <div class="storages-page">
    <!-- 头部区域 -->
    <div class="header">
      <div class="header-left">
        <n-input
          v-model:value="searchKeyword"
          placeholder="请输入路径搜索"
          clearable
          style="width: 240px; margin-right: 12px"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <n-icon :size="16" :depth="3">
              <SearchOutline />
            </n-icon>
          </template>
        </n-input>
        <n-button type="primary" @click="handleSearch" style="margin-right: 8px">
          <template #icon>
            <n-icon>
              <SearchOutline />
            </n-icon>
          </template>
          搜索
        </n-button>
        <n-button @click="handleReset">
          <template #icon>
            <n-icon>
              <RefreshOutline />
            </n-icon>
          </template>
          重置
        </n-button>
      </div>
      <div class="header-right">
        <n-button type="primary" @click="showAddModal = true">
          <template #icon>
            <n-icon>
              <AddOutline />
            </n-icon>
          </template>
          新增挂载
        </n-button>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <n-spin size="large">
        <template #description>
          <n-text depth="2">正在加载存储数据...</n-text>
        </template>
      </n-spin>
    </div>

    <!-- 存储卡片列表 -->
    <div v-else class="storage-cards">
      <n-card
        v-for="storage in tableData"
        :key="storage.id"
        class="storage-card"
        hoverable
        :bordered="false"
      >
        <div class="card-header">
          <div class="storage-info">
            <div class="storage-title">
              <n-text strong class="storage-name">{{ storage.name || '未命名存储' }}</n-text>
            </div>
            <n-ellipsis class="storage-path" :tooltip="{ placement: 'top' }">
              {{ storage.fullPath || '-' }}
            </n-ellipsis>
          </div>
          <div class="storage-actions">
            <n-button size="small" quaternary circle @click="handleDelete(storage)">
              <template #icon>
                <n-icon :size="16">
                  <TrashOutline />
                </n-icon>
              </template>
            </n-button>
            <n-dropdown
              :options="getRefreshOptions(storage.id)"
              @select="handleRefreshSelect"
              trigger="click"
            >
              <n-button size="small" quaternary circle>
                <template #icon>
                  <n-icon :size="16">
                    <RefreshOutline />
                  </n-icon>
                </template>
              </n-button>
            </n-dropdown>
          </div>
        </div>

        <n-divider style="margin: 16px 0" />

        <div class="card-content">
          <div class="info-grid">
            <div class="info-item">
              <div class="info-label">
                <n-icon :size="14" class="info-icon">
                  <FolderOutline />
                </n-icon>
                <span>存储类型</span>
              </div>
              <n-tag :color="getOsTypeColor(storage.osType)" size="small" class="info-tag">
                {{ getOsTypeDisplayName(storage.osType) }}
              </n-tag>
            </div>

            <div class="info-item">
              <div class="info-label">
                <n-icon :size="14" class="info-icon">
                  <KeyOutline />
                </n-icon>
                <span>绑定令牌</span>
              </div>
              <n-text class="info-value">{{ storage.tokenName || '未绑定' }}</n-text>
            </div>

            <div class="info-item full-width">
              <div class="info-label">
                <n-icon :size="14" class="info-icon">
                  <RefreshOutline />
                </n-icon>
                <span>自动刷新</span>
              </div>
              <div class="refresh-info">
                <n-tag :type="storage.enableAutoRefresh ? 'success' : 'default'" size="small">
                  {{ storage.enableAutoRefresh ? '已启用' : '未启用' }}
                </n-tag>
                <div v-if="storage.enableAutoRefresh" class="refresh-details">
                  <n-text depth="3" class="refresh-detail">
                    {{ storage.refreshInterval }}分钟间隔
                  </n-text>
                  <n-text depth="3" class="refresh-detail">
                    {{ storage.enableDeepRefresh ? '深度刷新' : '普通刷新' }}
                  </n-text>
                </div>
              </div>
            </div>

            <div class="info-item full-width">
              <div class="info-label">
                <n-icon :size="14" class="info-icon">
                  <TimeOutline />
                </n-icon>
                <span>更新时间</span>
              </div>
              <n-text class="info-value">{{ formatDateTime(storage.updatedAt) }}</n-text>
            </div>
          </div>
        </div>
      </n-card>

      <!-- 空状态 -->
      <div v-if="tableData.length === 0" class="empty-state">
        <n-empty description="暂无存储数据" size="large">
          <template #icon>
            <n-icon size="64" :depth="3">
              <ServerOutline />
            </n-icon>
          </template>
          <template #extra>
            <n-text depth="3">您还没有配置任何存储挂载点</n-text>
          </template>
        </n-empty>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="!loading && tableData.length > 0" class="pagination-container">
      <n-pagination
        v-model:page="paginationReactive.page"
        v-model:page-size="paginationReactive.pageSize"
        :item-count="paginationReactive.itemCount"
        :page-sizes="paginationReactive.pageSizes"
        show-size-picker
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </div>

    <!-- 新增挂载类型选择弹窗 -->
    <n-modal v-model:show="showAddModal" preset="dialog" title="请选择挂载类型">
      <template #header>
        <div style="display: flex; align-items: center; gap: 8px">
          <span>请选择挂载类型</span>
        </div>
      </template>
      <div class="mount-type-selection">
        <div class="mount-type-grid">
          <div
            v-for="mountType in mountTypes"
            :key="mountType.value"
            class="mount-type-card"
            @click="handleSelectMountType(mountType.value)"
          >
            <div class="mount-type-icon">
              <n-icon :size="32" :color="mountType.color">
                <component :is="mountType.icon" />
              </n-icon>
            </div>
            <div class="mount-type-info">
              <n-text strong class="mount-type-title">{{ mountType.label }}</n-text>
              <n-text depth="3" class="mount-type-desc">{{ mountType.description }}</n-text>
            </div>
          </div>
        </div>
      </div>
      <template #action>
        <n-button @click="showAddModal = false">取消</n-button>
      </template>
    </n-modal>

    <!-- 订阅号挂载弹窗 -->
    <SubscribeMountModal
      v-model:show="showSubscribeMountModal"
      @confirm="handleSubscribeMountConfirm"
    />

    <!-- 文件分享挂载弹窗 -->
    <ShareMountModal v-model:show="showShareMountModal" @confirm="handleShareMountConfirm" />

    <!-- 个人文件夹挂载弹窗 -->
    <PersonMountModal v-model:show="showPersonMountModal" @confirm="handlePersonMountConfirm" />

    <!-- 家庭文件夹挂载弹窗 -->
    <FamilyMountModal v-model:show="showFamilyMountModal" @confirm="handleFamilyMountConfirm" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  NInput,
  NButton,
  NCard,
  NText,
  NTag,
  NSpin,
  NEmpty,
  NPagination,
  NDivider,
  NEllipsis,
  NIcon,
  NModal,
  NDropdown,
  useMessage,
  useDialog,
  type PaginationProps,
  type DropdownOption,
} from 'naive-ui'
import {
  SearchOutline,
  RefreshOutline,
  FolderOutline,
  KeyOutline,
  TimeOutline,
  ServerOutline,
  AddOutline,
  TrashOutline,
} from '@vicons/ionicons5'
import { getStorageList, refreshStorage, deleteStorage } from '@/api/storage'
import type { StorageInfo } from '@/api/storage'
import { formatDateTime } from '@/utils/time'
import { getOsTypeDisplayName, getOsTypeColor, mountTypeConfigs } from '@/utils/osType'
import {
  SubscribeMountModal,
  ShareMountModal,
  PersonMountModal,
  FamilyMountModal,
} from '@/components/storage'

// 表格数据
const tableData = ref<StorageInfo[]>([])
const loading = ref(false)
const searchKeyword = ref('')

// 弹窗控制
const showAddModal = ref(false)
const showSubscribeMountModal = ref(false)
const showShareMountModal = ref(false)
const showPersonMountModal = ref(false)
const showFamilyMountModal = ref(false)

// 消息提示和对话框
const message = useMessage()
const dialog = useDialog()

// 分页配置
const paginationReactive = reactive<PaginationProps>({
  page: 1,
  pageSize: 12, // 改为12，适合3x4或4x3的网格布局
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [12, 24, 48, 96],
})

// 分页处理函数
const handlePageChange = (page: number) => {
  console.log('分页切换到:', page)
  paginationReactive.page = page
  fetchStorageList()
}

const handlePageSizeChange = (pageSize: number) => {
  console.log('每页大小切换到:', pageSize)
  paginationReactive.pageSize = pageSize
  paginationReactive.page = 1
  fetchStorageList()
}

// 获取存储列表
const fetchStorageList = () => {
  loading.value = true

  const params = {
    currentPage: paginationReactive.page || 1,
    pageSize: paginationReactive.pageSize || 10,
    path: searchKeyword.value || undefined,
  }

  console.log('请求参数:', params)

  getStorageList(params)
    .then((response) => {
      console.log('API响应:', response)

      if (response.code === 200 && response.data) {
        tableData.value = response.data.data || []
        paginationReactive.itemCount = response.data.total || 0
        console.log('表格数据:', tableData.value)
        console.log('总数据量:', paginationReactive.itemCount)
      }
    })
    .catch((error) => {
      console.error('获取存储列表失败:', error)
    })
    .finally(() => {
      loading.value = false
    })
}

// 搜索
const handleSearch = () => {
  paginationReactive.page = 1 // 搜索时重置到第一页
  fetchStorageList()
  console.log('搜索关键词:', searchKeyword.value)
}

// 重置
const handleReset = () => {
  searchKeyword.value = ''
  paginationReactive.page = 1 // 重置时回到第一页
  fetchStorageList()
}

// 挂载类型配置
const mountTypes = mountTypeConfigs

// 选择挂载类型
const handleSelectMountType = (mountType: string) => {
  console.log('选择的挂载类型:', mountType)
  showAddModal.value = false

  if (mountType === 'subscribe') {
    // 打开订阅号挂载弹窗
    showSubscribeMountModal.value = true
  } else if (mountType === 'share_folder') {
    // 打开文件分享挂载弹窗
    showShareMountModal.value = true
  } else if (mountType === 'person_folder') {
    // 打开个人文件夹挂载弹窗
    showPersonMountModal.value = true
  } else if (mountType === 'family_folder') {
    // 打开家庭文件夹挂载弹窗
    showFamilyMountModal.value = true
  } else {
    // 其他类型暂时显示提示
    message.info(
      `您选择了：${mountTypes.find((t) => t.value === mountType)?.label}，该功能正在开发中`
    )
  }
}

// 处理订阅号挂载确认
const handleSubscribeMountConfirm = (data: unknown) => {
  console.log('订阅号挂载数据:', data)

  // 检查是否是成功回调
  if (typeof data === 'object' && data !== null && 'success' in data) {
    message.success('挂载点创建成功')
    // 刷新列表
    fetchStorageList()
  } else {
    console.log('其他类型的回调数据:', data)
  }
}

// 处理文件分享挂载确认
const handleShareMountConfirm = (data: unknown) => {
  console.log('文件分享挂载数据:', data)

  // 检查是否是成功回调
  if (typeof data === 'object' && data !== null && 'success' in data) {
    message.success('挂载点创建成功')
    // 刷新列表
    fetchStorageList()
  } else {
    console.log('其他类型的回调数据:', data)
  }
}

// 处理个人文件夹挂载确认
const handlePersonMountConfirm = (data: unknown) => {
  console.log('个人文件夹挂载数据:', data)

  // 检查是否是成功回调
  if (typeof data === 'object' && data !== null && 'success' in data) {
    message.success('挂载点创建成功')
    // 刷新列表
    fetchStorageList()
  } else {
    console.log('其他类型的回调数据:', data)
  }
}

// 处理家庭文件夹挂载确认
const handleFamilyMountConfirm = (data: unknown) => {
  console.log('家庭文件夹挂载数据:', data)

  // 检查是否是成功回调
  if (typeof data === 'object' && data !== null && 'success' in data) {
    message.success('挂载点创建成功')
    // 刷新列表
    fetchStorageList()
  } else {
    console.log('其他类型的回调数据:', data)
  }
}

// 获取刷新选项
const getRefreshOptions = (storageId: number): DropdownOption[] => {
  return [
    {
      label: '普通刷新',
      key: `normal-${storageId}`,
      props: {
        onClick: () => handleRefresh(storageId, false),
      },
    },
    {
      label: '深度刷新',
      key: `deep-${storageId}`,
      props: {
        onClick: () => handleRefresh(storageId, true),
      },
    },
  ]
}

// 处理刷新选择
const handleRefreshSelect = (key: string) => {
  // 这个函数实际上不会被调用，因为我们使用了 props.onClick
  console.log('刷新选择:', key)
}

// 处理刷新
const handleRefresh = (storageId: number, deep: boolean) => {
  const storage = tableData.value.find((s) => s.id === storageId)
  const refreshType = deep ? '深度刷新' : '普通刷新'

  message.loading(`正在执行${refreshType}...`)

  refreshStorage({ id: storageId, deep })
    .then((response) => {
      if (response.code === 200) {
        message.success(`${storage?.name || '存储'} ${refreshType}成功`)
        // 刷新列表
        fetchStorageList()
      } else {
        message.error(
          `${refreshType}失败: ${(response as { message?: string }).message || '未知错误'}`
        )
      }
    })
    .catch((error) => {
      console.error('刷新存储失败:', error)
      message.error(`刷新失败: ${error instanceof Error ? error.message : '网络错误'}`)
    })
    .finally(() => {
      // 可以在这里添加清理逻辑
    })
}

// 处理删除
const handleDelete = (storage: StorageInfo) => {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除存储挂载点 "${storage.name || '未命名存储'}" 吗？此操作不可撤销。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: () => {
      message.loading(`正在删除 ${storage.name || '存储'}...`)

      deleteStorage({ id: storage.id })
        .then((response) => {
          if (response.code === 200) {
            message.success(`${storage.name || '存储'} 删除成功`)
            // 刷新列表
            fetchStorageList()
          } else {
            message.error(`删除失败: ${(response as { message?: string }).message || '未知错误'}`)
          }
        })
        .catch((error) => {
          console.error('删除存储失败:', error)
          message.error(`删除失败: ${error instanceof Error ? error.message : '网络错误'}`)
        })
        .finally(() => {
          // 可以在这里添加清理逻辑
        })
    },
  })
}

// 初始化
onMounted(() => {
  console.log('页面挂载，开始获取数据')
  fetchStorageList()
})
</script>

<style scoped>
/* 页面整体样式 */
.storages-page {
  padding: 24px;
  background: var(--n-color-target);
  flex: 1;
}

/* 头部搜索区域 */
.header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: var(--n-card-color);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgb(0 0 0 / 6%);
  border: 1px solid var(--n-border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right {
  flex-shrink: 0;
}

/* 加载状态 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  background: var(--n-card-color);
  border-radius: 12px;
  border: 1px solid var(--n-border-color);
  box-shadow: 0 2px 8px rgb(0 0 0 / 6%);
}

/* 存储卡片网格布局 */
.storage-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

/* 单个存储卡片样式 */
.storage-card {
  background: var(--n-card-color);
  border-radius: 12px;
  border: 1px solid var(--n-border-color);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  box-shadow: 0 2px 8px rgb(0 0 0 / 4%);
}

.storage-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgb(0 0 0 / 15%);
  border-color: var(--n-primary-color);
}

/* 卡片头部 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 0;
  margin-bottom: 0;
}

.storage-info {
  flex: 1;
  min-width: 0;

  /* 确保flex子项可以收缩 */
}

.storage-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.storage-name {
  font-size: 18px;
  font-weight: 600;
  color: var(--n-text-color);
  line-height: 1.4;

  /* 单行显示，超出省略号 */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;

  /* 确保为操作按钮留出空间 */
  max-width: calc(100% - 80px);
}

.storage-path {
  font-size: 13px;
  color: var(--n-text-color-2);
  line-height: 1.4;

  /* 最多两行显示，超出省略号 */
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
}

.storage-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
  align-items: flex-start;
}

.storage-actions .n-button {
  background-color: var(--n-color-target);
  border: 1px solid var(--n-border-color);
  transition: all 0.3s ease;
  width: 32px;
  height: 32px;
}

.storage-actions .n-button:hover {
  border-color: var(--n-primary-color);
  transform: scale(1.1);
}

.storage-actions .n-button:hover .n-icon {
  color: var(--n-primary-color);
}

/* 卡片内容 */
.card-content {
  padding: 0;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item.full-width {
  grid-column: 1 / -1;
}

.info-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: var(--n-text-color-2);
}

.info-icon {
  color: var(--n-text-color-3);
  flex-shrink: 0;
}

.info-value {
  font-size: 14px;
  color: var(--n-text-color);
  font-weight: 500;
}

.info-tag {
  align-self: flex-start;
  font-weight: 500;
}

/* 刷新信息样式 */
.refresh-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.refresh-details {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.refresh-detail {
  font-size: 12px;
  color: var(--n-text-color-3);
  background: var(--n-color-target);
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 500;
  border: 1px solid var(--n-border-color);
}

/* 空状态 */
.empty-state {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  background: var(--n-card-color);
  border-radius: 12px;
  border: 2px dashed var(--n-border-color);
}

/* 分页容器 */
.pagination-container {
  display: flex;
  justify-content: center;
  padding: 20px;
  background: var(--n-card-color);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgb(0 0 0 / 6%);
  border: 1px solid var(--n-border-color);
}

/* 响应式设计 */
@media (width <=1400px) {
  .storage-cards {
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 18px;
  }
}

@media (width <=1200px) {
  .storage-cards {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }
}

@media (width <=768px) {
  .storages-page {
    padding: 16px;
  }

  .storage-cards {
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 14px;
  }

  .header {
    padding: 16px;
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .header-left {
    justify-content: center;
  }

  .header-right {
    align-self: center;
    display: flex;
    justify-content: flex-end;
    width: 100%;
  }

  .storage-name {
    max-width: calc(100% - 70px);
  }
}

@media (width <=480px) {
  .storages-page {
    padding: 12px;
  }

  .storage-cards {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .header {
    padding: 12px;
  }

  .pagination-container {
    padding: 12px;
  }

  .info-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .info-item.full-width {
    grid-column: 1;
  }

  .storage-name {
    max-width: calc(100% - 60px);
  }

  .storage-actions {
    gap: 4px;
  }

  .storage-actions .n-button {
    width: 28px;
    height: 28px;
  }
}

/* 挂载类型选择弹窗样式 */
.mount-type-selection {
  padding: 16px 0;
}

.mount-type-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.mount-type-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: var(--n-card-color);
}

.mount-type-card:hover {
  border-color: var(--n-primary-color);
  background: var(--n-color-target);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgb(0 0 0 / 10%);
}

.mount-type-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background: var(--n-color-target);
}

.mount-type-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mount-type-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--n-text-color);
}

.mount-type-desc {
  font-size: 13px;
  color: var(--n-text-color-3);
  line-height: 1.4;
}

/* 弹窗响应式设计 */
@media (width >=768px) {
  .mount-type-grid {
    gap: 16px;
  }

  .mount-type-card {
    padding: 20px;
  }

  .mount-type-icon {
    width: 56px;
    height: 56px;
  }
}
</style>
