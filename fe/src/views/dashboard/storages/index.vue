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
              <n-tag size="small" type="info" class="storage-id-tag"> ID: {{ storage.id }} </n-tag>
            </div>
            <n-ellipsis class="storage-path" :tooltip="{ placement: 'top' }">
              {{ storage.fullPath || '-' }}
            </n-ellipsis>
          </div>
          <div class="storage-actions">
            <!-- 操作按钮区域，暂时留空 -->
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
  useMessage,
  type PaginationProps,
} from 'naive-ui'
import {
  SearchOutline,
  RefreshOutline,
  FolderOutline,
  KeyOutline,
  TimeOutline,
  ServerOutline,
  AddOutline,
} from '@vicons/ionicons5'
import { getStorageList } from '@/api/storage'
import type { StorageInfo } from '@/api/storage'
import { formatDateTime } from '@/utils/time'
import { getOsTypeDisplayName, getOsTypeColor, mountTypeConfigs } from '@/utils/osType'
import { SubscribeMountModal } from '@/components/storage'

// 表格数据
const tableData = ref<StorageInfo[]>([])
const loading = ref(false)
const searchKeyword = ref('')

// 弹窗控制
const showAddModal = ref(false)
const showSubscribeMountModal = ref(false)

// 消息提示
const message = useMessage()

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

// 初始化
onMounted(() => {
  console.log('页面挂载，开始获取数据')
  fetchStorageList()
})
</script>

<style scoped>
.storages-page {
  padding: 0;
  background: var(--n-color-target);
}

.header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: var(--n-card-color);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgb(0 0 0 / 6%);
}

.header-left {
  display: flex;
  align-items: center;
}

/* 加载状态 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
  background: var(--n-card-color);
  border-radius: 8px;
  margin: 20px;
}

/* 卡片网格布局 */
.storage-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
  padding: 0 20px;
  margin-bottom: 24px;
}

/* 卡片样式 */
.storage-card {
  background: var(--n-card-color);
  border-radius: 12px;
  border: 1px solid var(--n-border-color);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
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
}

.storage-info {
  flex: 1;
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
}

.storage-id-tag {
  font-size: 11px;
  font-weight: 500;
}

.storage-path {
  font-size: 13px;
  color: var(--n-text-color-2);
  line-height: 1.4;
  max-width: 100%;
}

.storage-actions {
  display: flex;
  gap: 8px;
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
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

/* 空状态 */
.empty-state {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
  background: var(--n-card-color);
  border-radius: 12px;
  border: 2px dashed var(--n-border-color);
}

/* 分页容器 */
.pagination-container {
  display: flex;
  justify-content: center;
  margin: 24px 20px;
  padding: 20px;
  background: var(--n-card-color);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgb(0 0 0 / 6%);
}

/* 响应式设计 */
@media (width <= 1400px) {
  .storage-cards {
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 18px;
  }
}

@media (width <= 1200px) {
  .storage-cards {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }
}

@media (width <= 768px) {
  .storage-cards {
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 14px;
    padding: 0 16px;
  }

  .header {
    margin: 0 16px 20px;
    padding: 16px;
  }

  .pagination-container {
    margin: 20px 16px;
    padding: 16px;
  }
}

@media (width <= 480px) {
  .storage-cards {
    grid-template-columns: 1fr;
    gap: 12px;
    padding: 0 12px;
  }

  .header {
    margin: 0 12px 16px;
    padding: 12px;
  }

  .pagination-container {
    margin: 16px 12px;
    padding: 12px;
  }

  .info-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .info-item.full-width {
    grid-column: 1;
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

/* 弹窗响应式设计 - 保持一行一个的风格 */
@media (width >= 768px) {
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
