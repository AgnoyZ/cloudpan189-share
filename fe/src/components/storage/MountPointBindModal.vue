<template>
  <n-modal v-model:show="visible" preset="dialog" title="绑定挂载点" style="width: 1000px">
    <template #header>
      <div style="display: flex; align-items: center; gap: 8px">
        <n-icon :size="20">
          <FolderOutline />
        </n-icon>
        <span>绑定挂载点</span>
      </div>
    </template>

    <div class="mount-bind-content">
      <!-- 批量操作区域 -->
      <div class="batch-actions">
        <div class="batch-operations">
          <div class="batch-token-select">
            <n-text depth="2" style="margin-right: 8px">批量设置令牌：</n-text>
            <n-select
              v-model:value="batchState.selectedToken"
              :options="cloudTokenOptions"
              placeholder="选择要批量应用的令牌"
              clearable
              style="width: 180px; margin-right: 8px"
            />
            <n-button
              type="primary"
              size="small"
              :disabled="batchState.selectedToken === undefined"
              @click="handleBatchApplyToken"
            >
              一键应用
            </n-button>
          </div>

          <div class="batch-path-prefix">
            <n-text depth="2" style="margin-right: 8px">批量设置路径前缀：</n-text>
            <n-input
              v-model:value="batchState.pathPrefix"
              placeholder="请输入路径前缀（必须以/开头）"
              style="width: 180px; margin-right: 8px"
            />
            <n-button
              type="primary"
              size="small"
              :disabled="!hasValidPathPrefix"
              @click="handleBatchApplyPathPrefix"
            >
              一键应用
            </n-button>
          </div>
        </div>
      </div>

      <div class="table-container">
        <n-data-table
          :columns="columns"
          :data="tableData"
          :pagination="false"
          :bordered="false"
          size="small"
          class="mount-table"
        />
      </div>
    </div>

    <template #action>
      <div class="modal-actions">
        <n-button @click="handleCancel">取消</n-button>
        <n-button type="primary" :loading="state.submitLoading" @click="handleConfirm">
          确认挂载
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, h, watch } from 'vue'
import {
  NModal,
  NIcon,
  NButton,
  NDataTable,
  NInput,
  NSelect,
  NText,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'
import { FolderOutline } from '@vicons/ionicons5'
import { addStorage, type AddStorageRequest, type AddStorageResponse } from '@/api/storage'
import type { ApiResponse } from '@/utils/api'
import { getCloudTokenList } from '@/api/cloudtoken'
import { getOsTypeDisplayName, getOsTypeColor } from '@/utils/osType'

// Props
interface MountItem {
  name: string
  osType: string
  subscribeUser?: string
  shareCode?: string
  shareAccessCode?: string
  cloudToken?: number
  fileId?: string
  familyId?: string
}

interface Props {
  show: boolean
  items: MountItem[]
  defaultCloudToken?: number
}

// Emits
interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'success'): void
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

// 状态管理
const state = reactive({
  submitLoading: false,
  cloudTokens: [] as Models.CloudToken[],
})

// 批量操作状态
const batchState = reactive({
  selectedToken: undefined as number | undefined,
  pathPrefix: '/',
})

// 表格数据
interface TableRow extends MountItem {
  id: string
  localPath: string
  selectedCloudToken?: number
}

const tableData = ref<TableRow[]>([])

// 计算属性
const cloudTokenOptions = computed(() => [
  { label: '不绑定', value: undefined },
  ...state.cloudTokens.map((token) => ({
    label: token.name,
    value: token.id,
  })),
])

const hasValidPathPrefix = computed(
  () => batchState.pathPrefix && batchState.pathPrefix.startsWith('/')
)

const hasInvalidRows = computed(() => tableData.value.some((row) => !row.localPath.trim()))

// 表格列定义
const columns: DataTableColumns<TableRow> = [
  {
    title: '序号',
    key: 'index',
    width: 80,
    render: (_, index) => index + 1,
  },
  {
    title: '识别出的名称',
    key: 'name',
    width: 200,
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: '挂载类型',
    key: 'osType',
    width: 150,
    render: (row) => {
      const displayName = getOsTypeDisplayName(row.osType)
      const colorInfo = getOsTypeColor(row.osType)
      return h('span', { style: { color: colorInfo.textColor } }, displayName)
    },
  },
  {
    title: '挂载路径',
    key: 'localPath',
    width: 250,
    render: (row, index) => {
      return h(NInput, {
        value: row.localPath,
        placeholder: '请输入挂载路径',
        onUpdateValue: (value: string) => {
          tableData.value[index].localPath = value
        },
      })
    },
  },
  {
    title: '绑定令牌',
    key: 'selectedCloudToken',
    width: 200,
    render: (row, index) => {
      return h(NSelect, {
        value: row.selectedCloudToken,
        options: cloudTokenOptions.value,
        placeholder: '选择云盘令牌',
        clearable: true,
        onUpdateValue: (value: number | undefined) => {
          tableData.value[index].selectedCloudToken = value
        },
      })
    },
  },
]

// 初始化表格数据
const initTableData = () => {
  tableData.value = props.items.map(
    (item, index) =>
      ({
        ...item,
        id: `item_${index}`,
        localPath: `/${item.name}`,
        selectedCloudToken: props.defaultCloudToken,
      }) as TableRow
  )
}

// 获取云盘令牌列表
const fetchCloudTokens = async () => {
  try {
    const response = await getCloudTokenList({ currentPage: 1, pageSize: 100 })
    if (response.code === 200 && response.data) {
      state.cloudTokens = response.data.data || []
    }
  } catch (error) {
    console.error('获取云盘令牌列表失败:', error)
  }
}

// 批量应用令牌
const handleBatchApplyToken = () => {
  if (batchState.selectedToken === undefined) {
    message.warning('请选择要应用的令牌')
    return
  }

  // 将选中的令牌应用到所有行
  tableData.value.forEach((row) => {
    row.selectedCloudToken = batchState.selectedToken
  })

  message.success('已批量应用令牌设置')
}

// 批量应用路径前缀
const handleBatchApplyPathPrefix = () => {
  if (!hasValidPathPrefix.value) {
    message.warning('路径前缀必须以 / 开头')
    return
  }

  // 确保前缀以 / 结尾（如果不是单独的 /）
  let prefix = batchState.pathPrefix
  if (prefix !== '/' && !prefix.endsWith('/')) {
    prefix += '/'
  }

  // 将路径前缀应用到所有行
  tableData.value.forEach((row) => {
    if (prefix === '/') {
      row.localPath = `/${row.name}`
    } else {
      row.localPath = `${prefix}${row.name}`
    }
  })

  message.success('已批量应用路径前缀设置')
}

// 取消
const handleCancel = () => {
  visible.value = false
}

// 构建请求数据
const buildRequests = (): AddStorageRequest[] => {
  return tableData.value.map((row) => ({
    localPath: row.localPath.trim(),
    osType: row.osType as AddStorageRequest['osType'],
    cloudToken: row.selectedCloudToken,
    subscribeUser: row.subscribeUser,
    shareCode: row.shareCode,
    shareAccessCode: row.shareAccessCode,
    fileId: row.fileId,
    familyId: row.familyId,
  }))
}

// 处理批量挂载结果
const handleMountResults = (responses: ApiResponse<AddStorageResponse>[]) => {
  const successCount = responses.filter((res) => res.code === 200).length
  const failCount = responses.length - successCount

  if (failCount === 0) {
    message.success(`成功挂载 ${successCount} 个存储点`)
    emit('success')
    visible.value = false
  } else {
    message.warning(`成功挂载 ${successCount} 个，失败 ${failCount} 个`)
    if (successCount > 0) {
      emit('success')
    }
  }
}

// 确认挂载
const handleConfirm = async () => {
  // 验证数据
  if (hasInvalidRows.value) {
    message.warning('请填写所有挂载路径')
    return
  }

  state.submitLoading = true

  try {
    // 构建请求数据
    const requests = buildRequests()

    // 批量添加存储挂载
    const responses = await Promise.all(requests.map((request) => addStorage(request)))

    handleMountResults(responses)
  } catch (error) {
    console.error('批量挂载失败:', error)
    message.error('批量挂载失败')
  } finally {
    state.submitLoading = false
  }
}

// 重置状态
const resetState = () => {
  batchState.selectedToken = undefined
  batchState.pathPrefix = '/'
  tableData.value = []
}

// 监听弹窗显示状态
watch(
  () => props.show,
  (newShow) => {
    if (newShow && props.items.length > 0) {
      initTableData()
      fetchCloudTokens()
    } else if (!newShow) {
      resetState()
    }
  },
  { immediate: true }
)

// 监听items变化
watch(
  () => props.items,
  (newItems) => {
    if (props.show && newItems.length > 0) {
      initTableData()
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.mount-bind-content {
  padding: 16px 0;
}

.batch-actions {
  margin-bottom: 16px;
  padding: 12px 16px;
  background: var(--n-card-color);
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
}

.batch-operations {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 24px;
}

.batch-token-select,
.batch-path-prefix {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.table-container {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid var(--n-border-color);
  border-radius: 6px;
}

.mount-table {
  min-height: 200px;
}

/* 固定表格标题 */
:deep(.n-data-table-thead) {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--n-th-color);
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* 表格样式优化 */
:deep(.n-data-table-th) {
  background: var(--n-th-color);
  font-weight: 600;
}

:deep(.n-data-table-td) {
  padding: 12px 8px;
}

/* 响应式设计 */
@media (width <= 768px) {
  .modal-actions {
    flex-direction: column;
  }
}
</style>
