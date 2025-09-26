<template>
  <n-modal v-model:show="visible" preset="dialog" title="个人文件夹挂载" style="width: 900px">
    <template #header>
      <div style="display: flex; align-items: center; gap: 8px">
        <n-icon :size="20">
          <PersonOutline />
        </n-icon>
        <span>个人文件夹挂载</span>
      </div>
    </template>

    <div class="person-mount-content">
      <!-- 第一步：选择令牌 -->
      <div v-if="currentStep === 1" class="step-content">
        <div class="step-header">
          <n-text strong>选择云盘令牌</n-text>
          <n-text depth="3">请选择要使用的天翼云盘令牌</n-text>
        </div>

        <div class="token-section">
          <div v-if="tokenState.loading" class="loading-container">
            <n-spin size="large">
              <template #description>
                <n-text depth="2">正在加载令牌列表...</n-text>
              </template>
            </n-spin>
          </div>

          <div v-else-if="tokenState.tokens.length === 0" class="empty-state">
            <n-empty description="暂无可用令牌" size="large">
              <template #icon>
                <n-icon size="64" :depth="3">
                  <KeyOutline />
                </n-icon>
              </template>
              <template #extra>
                <n-text depth="3">请先添加天翼云盘令牌</n-text>
              </template>
            </n-empty>
          </div>

          <div v-else class="token-list">
            <div
              v-for="token in tokenState.tokens"
              :key="token.id"
              class="token-card"
              :class="{ active: tokenState.selectedTokenId === token.id }"
              @click="handleSelectToken(token.id)"
            >
              <div class="token-info">
                <div class="token-header">
                  <n-text strong class="token-name">{{ token.name || '未命名令牌' }}</n-text>
                  <n-tag size="small" type="info" class="token-id-tag"> ID: {{ token.id }} </n-tag>
                </div>
                <n-text depth="3" class="token-username">{{ token.username }}</n-text>
              </div>
              <div class="token-actions">
                <n-icon v-if="tokenState.selectedTokenId === token.id" :size="20" color="#18a058">
                  <CheckmarkCircleOutline />
                </n-icon>
              </div>
            </div>
          </div>

          <div v-if="tokenState.tokens.length > 0" class="step-actions">
            <n-button
              type="primary"
              size="large"
              :disabled="!tokenState.selectedTokenId"
              @click="handleNextToFileSelection"
              style="width: 100%"
            >
              <template #icon>
                <n-icon>
                  <ArrowForwardOutline />
                </n-icon>
              </template>
              下一步：选择文件夹
            </n-button>
          </div>
        </div>
      </div>

      <!-- 第二步：选择文件夹 -->
      <div v-if="currentStep === 2" class="step-content">
        <div class="step-header">
          <n-text strong>选择文件夹</n-text>
          <n-text depth="3">请选择要挂载的个人文件夹（只能选择文件夹类型）</n-text>
        </div>

        <div class="file-section">
          <div v-if="fileState.loading" class="loading-container">
            <n-spin size="large">
              <template #description>
                <n-text depth="2">正在加载文件列表...</n-text>
              </template>
            </n-spin>
          </div>

          <div v-else class="file-tree-container">
            <!-- 面包屑导航 -->
            <div class="breadcrumb-container">
              <n-breadcrumb>
                <n-breadcrumb-item @click="handleNavigateToRoot">
                  <n-icon :size="16">
                    <HomeOutline />
                  </n-icon>
                  根目录
                </n-breadcrumb-item>
                <n-breadcrumb-item
                  v-for="(item, index) in fileState.breadcrumbs"
                  :key="item.id"
                  @click="handleNavigateToBreadcrumb(index)"
                >
                  {{ item.name }}
                </n-breadcrumb-item>
              </n-breadcrumb>
            </div>

            <!-- 文件树 -->
            <div class="file-tree">
              <n-tree
                :data="fileTreeData"
                :render-label="renderTreeLabel"
                :render-prefix="renderTreePrefix"
                :render-suffix="renderTreeSuffix"
                :selected-keys="fileState.selectedKeys"
                :expanded-keys="fileState.expandedKeys"
                :loading="fileState.treeLoading"
                block-line
                selectable
                @update:selected-keys="handleTreeSelect"
                @update:expanded-keys="handleTreeExpand"
              />
            </div>

            <!-- 当前选择信息 -->
            <div v-if="fileState.selectedFile" class="selected-info">
              <div class="selected-card">
                <div class="selected-header">
                  <n-icon :size="20" color="#18a058">
                    <CheckmarkCircleOutline />
                  </n-icon>
                  <n-text strong>已选择文件夹</n-text>
                </div>
                <div class="selected-details">
                  <n-text class="selected-name">{{ fileState.selectedFile.name }}</n-text>
                  <n-text depth="3" class="selected-path">
                    路径：{{ getSelectedFilePath() }}
                  </n-text>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <template #action>
      <div class="modal-actions">
        <n-button v-if="currentStep === 2" @click="handleBackToTokenSelection">
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
          :disabled="!fileState.selectedFile"
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
    :default-cloud-token="tokenState.selectedTokenId || undefined"
    @success="handleMountBindSuccess"
  />
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import {
  NModal,
  NIcon,
  NText,
  NButton,
  NSpin,
  NEmpty,
  NTag,
  NBreadcrumb,
  NBreadcrumbItem,
  NTree,
  useMessage,
} from 'naive-ui'
import {
  PersonOutline,
  KeyOutline,
  CheckmarkCircleOutline,
  ArrowForwardOutline,
  ArrowBackOutline,
  HomeOutline,
  FolderOutline,
  DocumentOutline,
} from '@vicons/ionicons5'
import { h } from 'vue'
import { getCloudTokenList } from '@/api/cloudtoken'
import { getPersonFiles } from '@/api/storage/advance'
import type { FileNode, GetPersonFilesQuery } from '@/api/storage/advance'
import { OS_TYPES } from '@/utils/osType'
import MountPointBindModal from './MountPointBindModal.vue'

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

// 令牌状态
const tokenState = reactive({
  loading: false,
  tokens: [] as Models.CloudToken[],
  selectedTokenId: null as number | null,
})

// 文件状态
const fileState = reactive({
  loading: false,
  treeLoading: false,
  files: [] as FileNode[],
  selectedFileId: null as string | null,
  selectedFile: null as FileNode | null,
  currentParentId: '-11', // 根目录ID
  breadcrumbs: [] as Array<{ id: string; name: string }>,
  selectedKeys: [] as string[],
  expandedKeys: [] as string[],
  treeData: new Map<string, FileNode[]>(), // 缓存各级目录的文件数据
})

// 绑定挂载点Modal状态
const mountBindState = reactive({
  show: false,
  items: [] as Array<{
    name: string
    osType: string
    cloudToken: number
    disableSwitchCloudToken: boolean
    fileId: string
  }>,
})

// 获取令牌列表
const fetchTokenList = () => {
  tokenState.loading = true
  getCloudTokenList({ noPaginate: true })
    .then((response) => {
      if (response.code === 200 && response.data) {
        tokenState.tokens = response.data.data || []
      } else {
        message.error(response.msg || '获取令牌列表失败')
      }
    })
    .catch((error) => {
      console.error('获取令牌列表失败:', error)
      message.error('获取令牌列表失败')
    })
    .finally(() => {
      tokenState.loading = false
    })
}

// 选择令牌
const handleSelectToken = (tokenId: number) => {
  tokenState.selectedTokenId = tokenId
}

// 下一步到文件选择
const handleNextToFileSelection = () => {
  if (!tokenState.selectedTokenId) {
    message.warning('请选择令牌')
    return
  }
  currentStep.value = 2
  fetchPersonFiles()
}

// 获取个人文件列表
const fetchPersonFiles = (parentId: string = '-11') => {
  if (!tokenState.selectedTokenId) return

  // 根目录显示主加载状态，子目录显示树加载状态
  if (parentId === '-11') {
    fileState.loading = true
  } else {
    fileState.treeLoading = true
  }

  const params: GetPersonFilesQuery = {
    pageNum: 1,
    pageSize: 100,
    cloudToken: tokenState.selectedTokenId,
    parentId: parentId,
  }

  console.log('请求文件列表:', params)

  getPersonFiles(params)
    .then((response) => {
      console.log('文件列表响应:', response)
      if (response.code === 200 && response.data) {
        const files = response.data.data || []

        if (parentId === '-11') {
          // 根目录
          fileState.files = files
          fileState.currentParentId = parentId
        }

        // 缓存当前目录的文件数据
        fileState.treeData.set(parentId, files)
        console.log('缓存文件数据:', parentId, files)
      } else {
        message.error(response.msg || '获取文件列表失败')
      }
    })
    .catch((error) => {
      console.error('获取文件列表失败:', error)
      message.error('获取文件列表失败')
    })
    .finally(() => {
      if (parentId === '-11') {
        fileState.loading = false
      } else {
        fileState.treeLoading = false
      }
    })
}

// 构建树形数据
const fileTreeData = computed(() => {
  const buildTreeNode = (file: FileNode): Record<string, unknown> => {
    const isFolder = file.isFolder === 1
    const hasChildren = isFolder && fileState.treeData.has(file.id)

    return {
      key: file.id,
      label: file.name,
      isLeaf: !isFolder,
      disabled: !isFolder, // 文件不可选择
      file: file,
      // 如果是文件夹但还没有加载子数据，设置为空数组以显示展开箭头
      children: isFolder
        ? hasChildren
          ? fileState.treeData.get(file.id)?.map(buildTreeNode)
          : []
        : undefined,
    }
  }

  return fileState.files.map(buildTreeNode)
})

// 树形组件渲染函数
const renderTreeLabel = ({ option }: { option: Record<string, unknown> }) => {
  return h('span', { class: 'tree-label' }, option.label as string)
}

const renderTreePrefix = ({ option }: { option: Record<string, unknown> }) => {
  const file = option.file as FileNode
  const isFolder = file.isFolder === 1
  return h(
    NIcon,
    {
      size: 18,
      color: isFolder ? '#ff9800' : '#2196f3',
    },
    {
      default: () => h(isFolder ? FolderOutline : DocumentOutline),
    }
  )
}

const renderTreeSuffix = ({ option }: { option: Record<string, unknown> }) => {
  const isSelected = fileState.selectedKeys.includes(option.key as string)
  const file = option.file as FileNode
  const isFolder = file.isFolder === 1

  if (isSelected && isFolder) {
    return h(
      NIcon,
      {
        size: 16,
        color: '#18a058',
      },
      {
        default: () => h(CheckmarkCircleOutline),
      }
    )
  }

  return null
}

// 树形选择处理
const handleTreeSelect = (keys: string[]) => {
  console.log('树形选择:', keys)
  fileState.selectedKeys = keys

  if (keys.length > 0) {
    const selectedKey = keys[0]

    // 递归查找选中的文件
    const findFileById = (files: FileNode[], id: string): FileNode | null => {
      for (const file of files) {
        if (file.id === id) return file
        // 如果有子节点，递归查找
        const childFiles = fileState.treeData.get(file.id)
        if (childFiles) {
          const found = findFileById(childFiles, id)
          if (found) return found
        }
      }
      return null
    }

    // 从所有缓存的数据中查找
    let selectedFile: FileNode | null = null
    for (const [, files] of fileState.treeData) {
      selectedFile = findFileById(files, selectedKey)
      if (selectedFile) break
    }

    // 如果在缓存中没找到，从当前文件列表中查找
    if (!selectedFile) {
      selectedFile = findFileById(fileState.files, selectedKey)
    }

    if (selectedFile && selectedFile.isFolder === 1) {
      fileState.selectedFile = selectedFile
      fileState.selectedFileId = selectedFile.id
      console.log('选中文件夹:', selectedFile)
    }
  } else {
    fileState.selectedFile = null
    fileState.selectedFileId = null
  }
}

// 树形展开处理
const handleTreeExpand = (keys: string[]) => {
  const newExpandedKeys = keys.filter((key) => !fileState.expandedKeys.includes(key))
  fileState.expandedKeys = keys

  // 加载新展开节点的子数据
  newExpandedKeys.forEach((key) => {
    if (!fileState.treeData.has(key)) {
      console.log('加载子目录:', key)
      fetchPersonFiles(key)
    }
  })
}

// 导航到根目录
const handleNavigateToRoot = () => {
  fileState.breadcrumbs = []
  fileState.selectedFileId = null
  fileState.selectedFile = null
  fetchPersonFiles('-11')
}

// 导航到面包屑
const handleNavigateToBreadcrumb = (index: number) => {
  const targetBreadcrumb = fileState.breadcrumbs[index]
  fileState.breadcrumbs = fileState.breadcrumbs.slice(0, index + 1)
  fileState.selectedFileId = null
  fileState.selectedFile = null
  fetchPersonFiles(targetBreadcrumb.id)
}

// 获取选中文件的完整路径
const getSelectedFilePath = () => {
  if (!fileState.selectedFile) return ''

  const pathParts = ['根目录']
  pathParts.push(...fileState.breadcrumbs.map((b) => b.name))
  pathParts.push(fileState.selectedFile.name)

  return pathParts.join(' / ')
}

// 返回令牌选择
const handleBackToTokenSelection = () => {
  currentStep.value = 1
}

// 取消
const handleCancel = () => {
  visible.value = false
}

// 确认挂载
const handleConfirm = () => {
  if (!fileState.selectedFile || !tokenState.selectedTokenId) {
    message.warning('请选择文件夹和令牌')
    return
  }

  // 将选择信息转换为挂载项
  mountBindState.items = [
    {
      name: fileState.selectedFile.name,
      osType: OS_TYPES.PERSON_FOLDER,
      cloudToken: tokenState.selectedTokenId,
      disableSwitchCloudToken: true, // 个人文件夹禁止修改令牌
      fileId: fileState.selectedFile.id,
    },
  ]

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
  tokenState.selectedTokenId = null
  tokenState.tokens = []
  fileState.loading = false
  fileState.treeLoading = false
  fileState.files = []
  fileState.selectedFileId = null
  fileState.selectedFile = null
  fileState.currentParentId = '-11'
  fileState.breadcrumbs = []
  fileState.selectedKeys = []
  fileState.expandedKeys = []
  fileState.treeData.clear()
  mountBindState.show = false
  mountBindState.items = []
}

// 监听弹窗打开，加载令牌列表
watch(visible, (newVal) => {
  if (newVal) {
    fetchTokenList()
  } else {
    resetAllState()
  }
})
</script>

<style scoped>
.person-mount-content {
  padding: 16px 0;
}

.step-content {
  min-height: 400px;
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

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

/* 令牌选择样式 */
.token-section {
  max-width: 600px;
  margin: 0 auto;
}

.token-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 24px;
}

.token-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: var(--n-card-color);
}

.token-card:hover {
  border-color: var(--n-primary-color);
  background: var(--n-color-target);
}

.token-card.active {
  border-color: var(--n-primary-color);
  background: var(--n-primary-color-suppl);
}

.token-info {
  flex: 1;
}

.token-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.token-name {
  font-size: 16px;
}

.token-id-tag {
  font-size: 11px;
}

.token-username {
  font-size: 13px;
}

.step-actions {
  margin-top: 24px;
}

/* 文件选择样式 */
.file-section {
  max-width: 800px;
  margin: 0 auto;
}

.file-tree-container {
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  background: var(--n-card-color);
  overflow: hidden;
}

.breadcrumb-container {
  padding: 12px 16px;
  border-bottom: 1px solid var(--n-border-color);
  background: var(--n-color-target);
}

.file-tree {
  max-height: 400px;
  overflow-y: auto;
}

.empty-folder {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

.file-list {
  padding: 8px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.file-item:hover {
  background: var(--n-color-target);
}

.file-item.selected {
  background: var(--n-primary-color-suppl);
  border: 1px solid var(--n-primary-color);
}

.file-item.disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.file-item.disabled:hover {
  background: transparent;
}

.file-icon {
  flex-shrink: 0;
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  display: block;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 2px;
  word-break: break-all;
}

.file-type {
  font-size: 12px;
}

.file-actions {
  flex-shrink: 0;
}

/* 选中信息样式 */
.selected-info {
  margin-top: 16px;
}

.selected-card {
  padding: 16px;
  background: var(--n-success-color-suppl);
  border: 1px solid var(--n-success-color);
  border-radius: 8px;
}

.selected-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.selected-name {
  display: block;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
}

.selected-path {
  font-size: 13px;
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* 响应式设计 */
@media (width <= 768px) {
  .token-card {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .token-actions {
    align-self: flex-end;
  }

  .file-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .modal-actions {
    flex-direction: column;
  }
}
</style>
