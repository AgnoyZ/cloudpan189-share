<template>
  <n-modal v-model:show="visible" preset="dialog" title="绑定文件" style="width: 800px">
    <div class="bind-files-modal">
      <!-- 搜索区域 -->
      <div class="search-section">
        <n-space>
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索文件..."
            clearable
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <n-icon :component="SearchOutline" />
            </template>
          </n-input>
          <n-button type="primary" @click="handleSearch">搜索</n-button>
        </n-space>
      </div>

      <!-- 文件列表 -->
      <div class="file-list-section">
        <n-data-table
          :columns="columns"
          :data="fileList"
          :loading="loading"
          :pagination="pagination"
          :row-key="(row) => row.id"
          :checked-row-keys="selectedFileIds"
          @update:checked-row-keys="handleSelectionChange"
        />
      </div>

      <!-- 已选择的文件 -->
      <div class="selected-section" v-if="selectedFileIds.length > 0">
        <n-divider />
        <div class="selected-header">
          <span>已选择 {{ selectedFileIds.length }} 个文件</span>
          <n-button text type="error" @click="clearSelection">清空选择</n-button>
        </div>
        <div class="selected-files">
          <n-tag
            v-for="fileId in selectedFileIds"
            :key="fileId"
            closable
            @close="removeSelection(fileId)"
          >
            {{ getFileName(fileId) }}
          </n-tag>
        </div>
      </div>
    </div>

    <template #action>
      <n-space>
        <n-button @click="handleCancel">取消</n-button>
        <n-button
          type="primary"
          :loading="submitting"
          :disabled="selectedFileIds.length === 0"
          @click="handleConfirm"
        >
          确定绑定
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import {
  NModal,
  NInput,
  NButton,
  NSpace,
  NIcon,
  NDataTable,
  NDivider,
  NTag,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'
import { SearchOutline } from '@vicons/ionicons5'
import { searchFiles, type FileSearchQuery, type FileSearchItem } from '@/api/file'
import { batchBindFiles } from '@/api/usergroup'

interface Props {
  show: boolean
  groupId?: number
  groupName?: string
  initialFileIds?: number[]
}

interface Emits {
  (e: 'update:show', value: boolean): void
  (e: 'success'): void
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  groupId: 0,
  groupName: '',
  initialFileIds: () => [],
})

const emit = defineEmits<Emits>()
const message = useMessage()

// 响应式数据
const visible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value),
})

const searchKeyword = ref('')
const loading = ref(false)
const submitting = ref(false)
const fileList = ref<FileSearchItem[]>([])
const selectedFileIds = ref<number[]>([])

// 分页配置
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  onChange: (page: number) => {
    pagination.value.page = page
    handleSearch()
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.value.pageSize = pageSize
    pagination.value.page = 1
    handleSearch()
  },
})

// 表格列配置
const columns: DataTableColumns<FileSearchItem> = [
  {
    type: 'selection',
  },
  {
    title: '文件名',
    key: 'name',
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: '类型',
    key: 'isDir',
    width: 80,
    render: (row) => (row.isDir ? '文件夹' : '文件'),
  },
  {
    title: '路径',
    key: 'fullPath',
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: '大小',
    key: 'size',
    width: 100,
    render: (row) => (row.isDir ? '-' : formatFileSize(row.size)),
  },
]

// 监听 props 变化
watch(
  () => props.show,
  (newShow) => {
    if (newShow) {
      // 重置状态
      searchKeyword.value = ''
      selectedFileIds.value = [...props.initialFileIds]
      // 初始加载文件列表
      nextTick(() => {
        handleSearch()
      })
    }
  }
)

// 搜索文件
const handleSearch = async () => {
  if (!visible.value) return

  loading.value = true
  try {
    const params: FileSearchQuery = {
      keyword: searchKeyword.value || undefined,
      global: true,
      pageSize: pagination.value.pageSize,
      currentPage: pagination.value.page,
    }

    const response = await searchFiles(params)
    if (response.code === 200 && response.data) {
      fileList.value = response.data.data
      pagination.value.itemCount = response.data.total
    } else {
      message.error(response.msg || '搜索文件失败')
    }
  } catch (error) {
    console.error('搜索文件失败:', error)
    message.error('搜索文件失败')
  } finally {
    loading.value = false
  }
}

// 处理选择变化
const handleSelectionChange = (keys: Array<string | number>) => {
  selectedFileIds.value = keys.map((key) => Number(key))
}

// 清空选择
const clearSelection = () => {
  selectedFileIds.value = []
}

// 移除单个选择
const removeSelection = (fileId: number) => {
  const index = selectedFileIds.value.indexOf(fileId)
  if (index > -1) {
    selectedFileIds.value.splice(index, 1)
  }
}

// 获取文件名
const getFileName = (fileId: number) => {
  const file = fileList.value.find((f) => f.id === fileId)
  return file ? file.name : `文件ID: ${fileId}`
}

// 格式化文件大小
const formatFileSize = (size: number) => {
  if (size === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(size) / Math.log(k))
  return parseFloat((size / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 取消操作
const handleCancel = () => {
  visible.value = false
}

// 确认绑定
const handleConfirm = async () => {
  if (!props.groupId || selectedFileIds.value.length === 0) {
    message.warning('请选择要绑定的文件')
    return
  }

  submitting.value = true
  try {
    const response = await batchBindFiles({
      groupId: props.groupId,
      fileIds: selectedFileIds.value,
    })

    if (response.code === 200) {
      message.success('文件绑定成功')
      visible.value = false
      emit('success')
    } else {
      message.error(response.msg || '文件绑定失败')
    }
  } catch (error) {
    console.error('文件绑定失败:', error)
    message.error('文件绑定失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.bind-files-modal {
  max-height: 600px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.search-section {
  margin-bottom: 16px;
}

.file-list-section {
  flex: 1;
  min-height: 300px;
}

.selected-section {
  margin-top: 16px;
}

.selected-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 500;
}

.selected-files {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  max-height: 100px;
  overflow-y: auto;
}
</style>
