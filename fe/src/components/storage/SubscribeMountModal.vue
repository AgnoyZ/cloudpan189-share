<template>
  <n-modal v-model:show="visible" preset="dialog" title="订阅号资源挂载" style="width: 1000px">
    <template #header>
      <div style="display: flex; align-items: center; gap: 8px">
        <n-icon :size="20">
          <DocumentTextOutline />
        </n-icon>
        <span>订阅号资源挂载</span>
      </div>
    </template>

    <div class="subscribe-mount-content">
      <!-- 第一步：输入订阅用户ID -->
      <div v-if="currentStep === 1" class="step-content">
        <div class="step-header">
          <n-text strong>第一步：输入订阅用户ID</n-text>
          <n-text depth="3">请输入要订阅的天翼云盘用户ID</n-text>
        </div>

        <div class="input-section">
          <n-input
            v-model:value="searchState.subscribeUserId"
            placeholder="请输入订阅用户ID"
            clearable
            size="large"
            @keyup.enter="handleSearchUser"
          >
            <template #prefix>
              <n-icon :size="16">
                <PersonOutline />
              </n-icon>
            </template>
          </n-input>

          <n-button
            type="primary"
            size="large"
            :loading="searchState.loading"
            :disabled="!isValidSubscribeUserId"
            @click="handleSearchUser"
            style="margin-top: 16px; width: 100%"
          >
            <template #icon>
              <n-icon>
                <SearchOutline />
              </n-icon>
            </template>
            查询分享列表
          </n-button>
        </div>
      </div>

      <!-- 第二步：展示资源列表 -->
      <div v-if="currentStep === 2" class="step-content">
        <div class="step-header">
          <n-text strong>第二步：选择要挂载的资源</n-text>
          <div class="user-info-line">
            <n-text depth="3"
              >用户：{{ resourceState.userInfo?.name || searchState.subscribeUserId }}</n-text
            >
            <n-text v-if="hasSelectedResources" type="primary" class="selected-count">
              <n-icon :size="14" color="#2196f3" style="vertical-align: middle; margin-right: 4px">
                <CheckmarkCircleOutline />
              </n-icon>
              已选中 {{ resourceState.selected.length }} 个文件
            </n-text>
          </div>
        </div>

        <!-- 批量操作区域 -->
        <div class="batch-actions">
          <div class="batch-select-actions">
            <n-button size="small" @click="handleSelectAll" :disabled="!hasResourceList">
              全选
            </n-button>
            <n-button
              size="small"
              @click="handleSelectNone"
              :disabled="!hasSelectedResources"
              style="margin-left: 8px"
            >
              取消全选
            </n-button>
          </div>

          <div class="search-section">
            <n-input
              v-model:value="resourceState.searchKeyword"
              placeholder="搜索资源名称"
              clearable
              @keyup.enter="handleSearchResource"
              style="width: 200px"
            >
              <template #prefix>
                <n-icon :size="16">
                  <SearchOutline />
                </n-icon>
              </template>
            </n-input>
            <n-button type="primary" @click="handleSearchResource" style="margin-left: 8px">
              搜索
            </n-button>
            <n-button @click="handleResetResourceSearch" style="margin-left: 8px"> 重置 </n-button>
          </div>
        </div>

        <!-- 资源表格 -->
        <div class="table-container">
          <n-spin :show="resourceState.loading">
            <n-data-table
              v-if="hasResourceList"
              :columns="resourceColumns"
              :data="resourceState.list"
              :pagination="false"
              :bordered="false"
              size="small"
              class="resource-table"
              :row-key="(row: ShareResourceInfo) => row.id"
            />

            <n-empty v-else description="暂无资源数据" size="large" style="min-height: 200px">
              <template #icon>
                <n-icon size="48" :depth="3">
                  <DocumentTextOutline />
                </n-icon>
              </template>
            </n-empty>
          </n-spin>
        </div>

        <!-- 分页 -->
        <div v-if="hasResourceList" class="pagination-section">
          <n-pagination
            v-model:page="resourcePagination.page"
            v-model:page-size="resourcePagination.pageSize"
            :item-count="resourcePagination.itemCount"
            :page-sizes="PAGINATION_CONFIG.PAGE_SIZES"
            show-size-picker
            @update:page="handleResourcePageChange"
            @update:page-size="handleResourcePageSizeChange"
          />
        </div>
      </div>
    </div>

    <template #action>
      <div class="modal-actions">
        <n-button v-if="currentStep === 2" @click="handleBackToStep1">
          <template #icon>
            <n-icon>
              <ArrowBackOutline />
            </n-icon>
          </template>
          返回上一步
        </n-button>
        <n-button @click="handleCancel">取消</n-button>
        <n-button
          v-if="currentStep === 2"
          type="primary"
          :disabled="!hasSelectedResources"
          @click="handleConfirm"
        >
          绑定挂载点
        </n-button>
      </div>
    </template>
  </n-modal>

  <!-- 绑定挂载点Modal -->
  <MountPointBindModal
    v-model:show="mountBindState.show"
    :items="mountBindState.items"
    @success="handleMountBindSuccess"
  />
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, h } from 'vue'
import {
  NModal,
  NIcon,
  NText,
  NInput,
  NButton,
  NSpin,
  NEmpty,
  NPagination,
  NDataTable,
  NCheckbox,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'
import {
  DocumentTextOutline,
  PersonOutline,
  SearchOutline,
  FolderOutline,
  DocumentOutline,
  CheckmarkCircleOutline,
  ArrowBackOutline,
} from '@vicons/ionicons5'
import { getSubscribeUser } from '@/api/storage/advance'
import type { ShareResourceInfo, GetSubscribeUserResponse } from '@/api/storage/advance'
import type { ApiResponse } from '@/utils/api'
import { formatDateTime } from '@/utils/time'
import { OS_TYPES } from '@/utils/osType'
import MountPointBindModal from './MountPointBindModal.vue'

// 常量定义
const PAGINATION_CONFIG = {
  DEFAULT_PAGE_SIZE: 30,
  PAGE_SIZES: [20, 30, 50, 100] as number[],
  DEFAULT_PAGE: 1,
}

// Props
interface Props {
  show: boolean
}

// Emits
interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'confirm', data: { success: boolean }): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 消息提示
const message = useMessage()

// 双向绑定
const visible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value),
})

// 步骤控制
const currentStep = ref(1)

// 第一步：用户搜索状态
const searchState = reactive({
  subscribeUserId: '',
  loading: false,
})

// 第二步：资源列表状态
const resourceState = reactive({
  userInfo: null as GetSubscribeUserResponse | null,
  list: [] as ShareResourceInfo[],
  loading: false,
  searchKeyword: '',
  selected: [] as ShareResourceInfo[],
})

// 分页配置
const resourcePagination = reactive({
  page: PAGINATION_CONFIG.DEFAULT_PAGE,
  pageSize: PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
  itemCount: 0,
})

// 绑定挂载点Modal状态
const mountBindState = reactive({
  show: false,
  items: [] as Array<{
    name: string
    osType: string
    subscribeUser: string
    shareCode: string
  }>,
})

// 计算属性
const isValidSubscribeUserId = computed(() => searchState.subscribeUserId.trim().length > 0)
const hasSelectedResources = computed(() => resourceState.selected.length > 0)
const hasResourceList = computed(() => resourceState.list.length > 0)

// 表格列定义
const resourceColumns: DataTableColumns<ShareResourceInfo> = [
  {
    title: '选择',
    key: 'select',
    width: 80,
    render: (row) => {
      const isSelected = resourceState.selected.some((r) => r.id === row.id)
      return h(NCheckbox, {
        checked: isSelected,
        onUpdateChecked: () => handleSelectResource(row),
      })
    },
  },
  {
    title: '序号',
    key: 'index',
    width: 80,
    render: (_, index) => (resourcePagination.page - 1) * resourcePagination.pageSize + index + 1,
  },
  {
    title: '资源名称',
    key: 'name',
    width: 300,
    ellipsis: {
      tooltip: true,
    },
    render: (row) => {
      return h('div', { style: 'display: flex; align-items: center; gap: 8px;' }, [
        h(
          NIcon,
          {
            size: 20,
            color: row.isFolder ? '#ff9800' : '#2196f3',
          },
          {
            default: () => (row.isFolder ? h(FolderOutline) : h(DocumentOutline)),
          }
        ),
        h('span', { style: 'word-break: break-all;' }, row.name),
      ])
    },
  },
  {
    title: '类型',
    key: 'type',
    width: 100,
    render: (row) => (row.isFolder ? '文件夹' : '单文件'),
  },
  {
    title: '分享时间',
    key: 'shareTime',
    width: 180,
    render: (row) => formatDateTime(row.shareTime),
  },
]

// 处理API响应的公共方法
const handleApiResponse = (
  response: ApiResponse<GetSubscribeUserResponse>,
  isInitialSearch = false
) => {
  if (response.code === 200 && response.data) {
    resourceState.userInfo = response.data
    resourceState.list = response.data.data || []
    resourcePagination.itemCount = response.data.total || 0

    if (isInitialSearch) {
      resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
      resourcePagination.pageSize = PAGINATION_CONFIG.DEFAULT_PAGE_SIZE

      if (resourceState.list.length > 0) {
        currentStep.value = 2
        message.success(`找到 ${resourcePagination.itemCount} 个资源`)
      } else {
        message.warning('该用户暂无分享资源')
      }
    }
  } else {
    message.error(response.msg || '获取用户资源失败')
  }
}

// 搜索用户资源
const handleSearchUser = () => {
  if (!isValidSubscribeUserId.value) {
    message.warning('请输入订阅用户ID')
    return
  }

  searchState.loading = true

  return getSubscribeUser({
    subscribeUser: searchState.subscribeUserId.trim(),
    currentPage: PAGINATION_CONFIG.DEFAULT_PAGE,
    pageSize: PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
  })
    .then((response) => {
      handleApiResponse(response, true)
    })
    .catch((error) => {
      console.error('搜索用户资源失败:', error)
      message.error('搜索用户资源失败')
    })
    .finally(() => {
      searchState.loading = false
    })
}

// 搜索资源
const handleSearchResource = () => {
  resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
  fetchResourceList()
}

// 重置资源搜索
const handleResetResourceSearch = () => {
  resourceState.searchKeyword = ''
  resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
  fetchResourceList()
}

// 获取资源列表
const fetchResourceList = () => {
  if (!isValidSubscribeUserId.value) return

  resourceState.loading = true

  return getSubscribeUser({
    subscribeUser: searchState.subscribeUserId.trim(),
    name: resourceState.searchKeyword || undefined,
    currentPage: resourcePagination.page,
    pageSize: resourcePagination.pageSize,
  })
    .then((response) => {
      handleApiResponse(response)
    })
    .catch((error) => {
      console.error('获取资源列表失败:', error)
      message.error('获取资源列表失败')
    })
    .finally(() => {
      resourceState.loading = false
    })
}

// 分页处理
const handleResourcePageChange = (page: number) => {
  resourcePagination.page = page
  fetchResourceList()
}

const handleResourcePageSizeChange = (pageSize: number) => {
  resourcePagination.pageSize = pageSize
  resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
  fetchResourceList()
}

// 选择资源（多选）
const handleSelectResource = (resource: ShareResourceInfo) => {
  const index = resourceState.selected.findIndex((r) => r.id === resource.id)
  if (index > -1) {
    // 已选中，取消选择
    resourceState.selected.splice(index, 1)
  } else {
    // 未选中，添加选择
    resourceState.selected.push(resource)
  }
}

// 全选
const handleSelectAll = () => {
  resourceState.selected = [...resourceState.list]
  message.success(`已选中 ${resourceState.list.length} 个资源`)
}

// 取消全选
const handleSelectNone = () => {
  resourceState.selected = []
  message.success('已取消所有选择')
}

// 返回上一步
const handleBackToStep1 = () => {
  currentStep.value = 1
  resourceState.selected = []
}

// 取消
const handleCancel = () => {
  visible.value = false
}

// 确认挂载
const handleConfirm = () => {
  if (!hasSelectedResources.value) {
    message.warning('请选择要挂载的资源')
    return
  }

  // 将选中的资源转换为挂载项
  mountBindState.items = resourceState.selected.map((resource) => ({
    name: resource.name,
    osType: OS_TYPES.SUBSCRIBE_SHARE_FOLDER,
    subscribeUser: searchState.subscribeUserId.trim(),
    shareCode: resource.accessCode,
  }))

  // 打开绑定挂载点Modal
  mountBindState.show = true
}

// 绑定挂载点成功回调
const handleMountBindSuccess = () => {
  message.success('挂载点绑定成功')
  emit('confirm', { success: true })
  visible.value = false
}

// 重置所有状态
const resetAllState = () => {
  currentStep.value = 1
  searchState.subscribeUserId = ''
  searchState.loading = false
  resourceState.userInfo = null
  resourceState.list = []
  resourceState.loading = false
  resourceState.searchKeyword = ''
  resourceState.selected = []
  resourcePagination.page = PAGINATION_CONFIG.DEFAULT_PAGE
  resourcePagination.pageSize = PAGINATION_CONFIG.DEFAULT_PAGE_SIZE
  resourcePagination.itemCount = 0
  mountBindState.show = false
  mountBindState.items = []
}

// 监听弹窗关闭，重置状态
watch(visible, (newVal) => {
  if (!newVal) {
    resetAllState()
  }
})
</script>

<style scoped>
.subscribe-mount-content {
  padding: 16px 0;
}

.step-content {
  min-height: 300px;
}

.step-header {
  margin-bottom: 24px;
  text-align: center;
}

.step-header .n-text:first-child {
  display: block;
  margin-bottom: 8px;
  font-size: 18px;
}

.user-info-line {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
  line-height: 1;
}

.selected-count {
  display: flex;
  align-items: center;
  font-size: 14px;
  line-height: 1;
}

.input-section {
  max-width: 400px;
  margin: 0 auto;
}

.batch-actions {
  margin-bottom: 16px;
  padding: 12px 16px;
  background: var(--n-card-color);
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.batch-select-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.table-container {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
}

.resource-table {
  min-height: 200px;
}

/* 固定表格标题 */
:deep(.n-data-table-thead) {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--n-th-color);
}

/* 表格样式优化 */
:deep(.n-data-table-th) {
  background: var(--n-th-color);
  font-weight: 600;
}

:deep(.n-data-table-td) {
  padding: 12px 8px;
}

.pagination-section {
  display: flex;
  justify-content: center;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--n-border-color);
}

.selected-info {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  margin-bottom: 12px;
  background: var(--n-primary-color-suppl);
  border: 1px solid var(--n-primary-color);
  border-radius: 6px;
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* 响应式设计 */
@media (width <= 768px) {
  .search-section {
    flex-direction: column;
    gap: 8px;
  }

  .search-section .n-input {
    width: 100%;
  }

  .resource-info {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .resource-meta {
    flex-direction: column;
    gap: 4px;
  }

  .modal-actions {
    flex-direction: column;
  }
}
</style>
