<template>
  <div class="autoingest-page">
    <!-- 顶部 Tabs -->
    <n-tabs type="line" v-model:value="activeTab">
      <n-tab name="plans">计划管理</n-tab>
      <n-tab name="logs">运行日志</n-tab>
    </n-tabs>

    <!-- 头部区域（Plans） -->
    <div v-if="activeTab === 'plans'" class="header">
      <div class="header-left">
        <n-input
          v-model:value="planQuery.name"
          placeholder="按名称搜索计划"
          clearable
          style="width: 240px; margin-right: 12px"
          @keyup.enter="handlePlanSearch"
        >
          <template #prefix>
            <n-icon :size="16" :depth="3">
              <SearchOutline />
            </n-icon>
          </template>
        </n-input>
        <n-button type="primary" @click="handlePlanSearch" style="margin-right: 8px">
          <template #icon>
            <n-icon>
              <SearchOutline />
            </n-icon>
          </template>
          搜索
        </n-button>
        <n-button @click="handlePlanReset">
          <template #icon>
            <n-icon>
              <RefreshOutline />
            </n-icon>
          </template>
          重置
        </n-button>
      </div>
      <div class="header-right">
        <n-button type="primary" @click="showCreateModal = true">
          <template #icon>
            <n-icon>
              <AddOutline />
            </n-icon>
          </template>
          新建订阅计划
        </n-button>
      </div>
    </div>

    <!-- 头部区域（Logs） -->
    <div v-else class="header">
      <div class="header-left">
        <n-select
          v-model:value="logQuery.planId"
          :options="planOptions"
          placeholder="按计划筛选"
          clearable
          style="width: 220px; margin-right: 8px"
        />
        <n-select
          v-model:value="logQuery.level"
          :options="logLevelOptions"
          placeholder="日志级别"
          clearable
          style="width: 160px; margin-right: 8px"
        />
        <n-button type="primary" @click="handleLogFilter" style="margin-right: 8px">
          <template #icon>
            <n-icon>
              <SearchOutline />
            </n-icon>
          </template>
          筛选
        </n-button>
        <n-button @click="handleLogReset">
          <template #icon>
            <n-icon>
              <RefreshOutline />
            </n-icon>
          </template>
          重置
        </n-button>
      </div>
      <div class="header-right">
        <n-text depth="3">最近刷新：{{ refreshTime.format('YYYY-MM-DD HH:mm:ss') }}</n-text>
      </div>
    </div>

    <!-- 计划管理表格 -->
    <n-data-table
      v-if="activeTab === 'plans'"
      :columns="planColumns"
      :data="planTable"
      :loading="planLoading"
      :pagination="planPagination"
      class="autoingest-table"
      remote
    />

    <!-- 运行日志表格 -->
    <n-data-table
      v-else
      :columns="logColumns"
      :data="logTable"
      :loading="logLoading"
      :pagination="logPagination"
      class="autoingest-table"
      remote
    />

    <!-- 新建计划弹窗 -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      title="新建订阅计划"
      :mask-closable="false"
    >
      <n-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-placement="left"
        label-width="140"
      >
        <n-form-item label="计划名称" path="name">
          <n-input v-model:value="createForm.name" placeholder="例如：订阅计划A" />
        </n-form-item>

        <n-form-item label="挂载父目录" path="parentPath">
          <n-input v-model:value="createForm.parentPath" placeholder="/Movies" />
        </n-form-item>

        <n-form-item label="上传用户ID" path="upUserId">
          <n-input v-model:value="createForm.upUserId" placeholder="订阅用户ID" />
        </n-form-item>

        <n-form-item label="绑定令牌" path="cloudToken">
          <n-select
            v-model:value="createForm.cloudToken"
            :options="cloudTokenOptions"
            placeholder="可选"
            clearable
            filterable
          />
        </n-form-item>

        <n-form-item label="冲突处理策略" path="onConflict">
          <n-radio-group v-model:value="createForm.onConflict">
            <n-space>
              <n-radio
                v-for="opt in AUTO_INGEST_ON_CONFLICT_OPTIONS"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>

        <n-form-item label="自动入库间隔(分钟)" path="autoIngestInterval">
          <n-input-number
            v-model:value="createForm.autoIngestInterval"
            :min="AUTO_INGEST_INTERVAL_MIN"
            :max="REFRESH_INTERVAL_MAX"
          />
        </n-form-item>

        <n-form-item label="一键添加历史" path="oneClickAddHistory">
          <n-switch v-model:value="createForm.oneClickAddHistory" />
        </n-form-item>

        <n-divider title-placement="left">刷新策略（可选）</n-divider>

        <n-form-item label="启用自动刷新" path="refreshStrategy.enableAutoRefresh">
          <n-switch v-model:value="createForm.refreshStrategy.enableAutoRefresh" />
        </n-form-item>

        <template v-if="createForm.refreshStrategy.enableAutoRefresh">
          <n-form-item label="刷新间隔(分钟)" path="refreshStrategy.refreshInterval">
            <n-input-number
              v-model:value="createForm.refreshStrategy.refreshInterval"
              :min="REFRESH_INTERVAL_MIN"
              :max="REFRESH_INTERVAL_MAX"
            />
          </n-form-item>
          <n-form-item label="持续天数" path="refreshStrategy.autoRefreshDays">
            <n-input-number
              v-model:value="createForm.refreshStrategy.autoRefreshDays"
              :min="AUTO_REFRESH_DAYS_MIN"
              :max="AUTO_REFRESH_DAYS_MAX"
            />
          </n-form-item>
          <n-form-item label="深度刷新" path="refreshStrategy.enableDeepRefresh">
            <n-switch v-model:value="createForm.refreshStrategy.enableDeepRefresh" />
          </n-form-item>
        </template>
      </n-form>

      <template #action>
        <n-space>
          <n-button @click="showCreateModal = false">取消</n-button>
          <n-button type="primary" :loading="createSubmitting" @click="handleCreatePlan">
            确认创建
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, h, onMounted } from 'vue'
import {
  NDataTable,
  NButton,
  NIcon,
  NInput,
  NText,
  NModal,
  NForm,
  NFormItem,
  NInputNumber,
  NSelect,
  NSwitch,
  NRadioGroup,
  NRadio,
  NSpace,
  NDivider,
  NPopconfirm,
  NTabs,
  NTab,
  NTag,
  useMessage,
  type DataTableColumns,
  type FormInst,
  type FormRules,
  type PaginationProps,
} from 'naive-ui'
import {
  SearchOutline,
  RefreshOutline,
  AddOutline,
  TrashOutline,
  CheckmarkCircleOutline,
  CloseCircleOutline,
} from '@vicons/ionicons5'
import {
  getAutoIngestPlanList,
  createSubscribePlan,
  enableAutoIngestPlan,
  disableAutoIngestPlan,
  deleteAutoIngestPlan,
  getAutoIngestLogList,
  type CreateSubscribePlanRequest,
} from '@/api/autoingest'
import { getCloudTokenList } from '@/api/cloudtoken'
import dayjs from 'dayjs'
import {
  AUTO_INGEST_ON_CONFLICT_OPTIONS,
  AUTO_INGEST_SOURCE_TYPE_OPTIONS,
  AUTO_INGEST_INTERVAL_MIN,
  REFRESH_INTERVAL_MIN,
  REFRESH_INTERVAL_MAX,
  AUTO_REFRESH_DAYS_MIN,
  AUTO_REFRESH_DAYS_MAX,
} from '@/constants/autoIngest'
import { type ApiResponse } from '@/utils/api'

const message = useMessage()

// Tabs
const activeTab = ref<'plans' | 'logs'>('plans')

// Refresh Time
const refreshTime = ref(dayjs())

// Cloud Token options
const cloudTokenOptions = ref<{ label: string; value: number }[]>([])
const loadCloudTokens = () => {
  getCloudTokenList({ noPaginate: true })
    .then((res: ApiResponse<Models.PaginationResponse<Models.CloudToken>>) => {
      if (res.code === 200 && res.data) {
        cloudTokenOptions.value = res.data.data.map((t) => ({
          label: t.name || `令牌${t.id}`,
          value: t.id,
        }))
      }
    })
    .catch((err: unknown) => {
      // 忽略错误，仅防止类型告警
      console.error('获取令牌列表失败:', err)
    })
}

// -------- Plans --------
const planLoading = ref(false)
const planTable = ref<Models.AutoIngestPlan[]>([])
const planQuery = reactive({
  name: '' as string | undefined,
})

// 分页（对齐用户组管理）
const planPagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  prefix: ({ itemCount }) => `共 ${itemCount} 条`,
  onChange: (page: number) => {
    planPagination.page = page
    fetchPlanList()
  },
  onUpdatePageSize: (ps: number) => {
    planPagination.pageSize = ps
    planPagination.page = 1
    fetchPlanList()
  },
})

const handlePlanSearch = () => {
  planPagination.page = 1
  fetchPlanList()
}

const handlePlanReset = () => {
  planQuery.name = ''
  planPagination.page = 1
  fetchPlanList()
}

const planColumns: DataTableColumns<Models.AutoIngestPlan> = [
  {
    title: '计划名称',
    key: 'name',
    width: 120,
    align: 'center',
    ellipsis: { tooltip: true },
  },
  {
    title: '来源类型',
    key: 'sourceType',
    width: 80,
    align: 'center',
    render: (row) => {
      const opt = AUTO_INGEST_SOURCE_TYPE_OPTIONS.find((o) => o.value === row.sourceType)
      return h(
        NTag,
        { type: 'primary', size: 'small', bordered: true },
        { default: () => opt?.label || '-' }
      )
    },
  },
  {
    title: '父目录',
    key: 'parentPath',
    width: 120,
    align: 'center',
    ellipsis: { tooltip: true },
    render: (row) =>
      h(
        'div',
        {
          style: 'max-width:240px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;',
          title: row.parentPath || '-',
        },
        row.parentPath || '-'
      ),
  },
  {
    title: '间隔(分钟)',
    key: 'autoIngestInterval',
    width: 90,
    align: 'center',
  },
  {
    title: '启用',
    key: 'enabled',
    width: 90,
    align: 'center',
    render: (row) => {
      const type = row.enabled ? 'success' : 'default'
      const label = row.enabled ? '启用' : '停用'
      return h(NTag, { type, size: 'small', bordered: true }, { default: () => label })
    },
  },
  {
    title: '统计',
    key: 'stats',
    width: 100,
    align: 'center',
    render: (row) =>
      h(
        'div',
        {
          class: 'stats-cell',
          style: 'display:flex;justify-content:center;gap:6px;align-items:center;',
        },
        [
          h(
            NTag,
            { type: 'success', size: 'small', bordered: true },
            { default: () => String(row.addCount || 0) }
          ),
          h('span', null, '/'),
          h(
            NTag,
            { type: 'error', size: 'small', bordered: true },
            { default: () => String(row.failedCount || 0) }
          ),
        ]
      ),
  },
  {
    title: '时间',
    key: 'time',
    width: 220,
    align: 'center',
    render: (row) =>
      h('div', { class: 'time-cell' }, [
        h('div', { class: 'time-line' }, [
          h('span', { class: 'time-label' }, '创建: '),
          h('span', { class: 'time-value' }, formatDT(row.createdAt)),
        ]),
        h('div', { class: 'time-line' }, [
          h('span', { class: 'time-label' }, '更新: '),
          h('span', { class: 'time-value' }, formatDT(row.updatedAt)),
        ]),
      ]),
  },
  {
    title: '操作',
    key: 'actions',
    minWidth: 180,
    align: 'center',
    render: (row) =>
      h(
        NSpace,
        { size: 'small', justify: 'center', align: 'center' },
        {
          default: () => [
            row.enabled
              ? h(
                  NButton,
                  { size: 'tiny', type: 'warning', secondary: true, onClick: () => onDisable(row) },
                  {
                    icon: () => h(NIcon, { size: 12 }, { default: () => h(CloseCircleOutline) }),
                    default: () => '停用',
                  }
                )
              : h(
                  NButton,
                  { size: 'tiny', type: 'success', secondary: true, onClick: () => onEnable(row) },
                  {
                    icon: () =>
                      h(NIcon, { size: 12 }, { default: () => h(CheckmarkCircleOutline) }),
                    default: () => '启用',
                  }
                ),
            h(
              NPopconfirm,
              {
                onPositiveClick: () => onDelete(row),
                negativeText: '取消',
                positiveText: '确认删除',
              },
              {
                trigger: () =>
                  h(
                    NButton,
                    { size: 'tiny', type: 'error', secondary: true },
                    {
                      icon: () => h(NIcon, { size: 12 }, { default: () => h(TrashOutline) }),
                      default: () => '删除',
                    }
                  ),
                default: () => `确定要删除计划 "${row.name || '#' + row.id}" 吗？此操作不可撤销。`,
              }
            ),
          ],
        }
      ),
  },
]

const fetchPlanList = () => {
  planLoading.value = true
  getAutoIngestPlanList({
    currentPage: planPagination.page || 1,
    pageSize: planPagination.pageSize || 10,
    name: planQuery.name || undefined,
  })
    .then((res: ApiResponse<Models.PaginationResponse<Models.AutoIngestPlan>>) => {
      if (res.code === 200 && res.data) {
        planTable.value = res.data.data
        planPagination.itemCount = res.data.total
        // 更新 planOptions 供日志筛选使用
        planOptions.value = [
          { label: '全部计划', value: undefined },
          ...res.data.data.map((p: Models.AutoIngestPlan) => ({
            label: `${p.name || '#' + p.id}`,
            value: p.id,
          })),
        ]
      }
    })
    .catch((err: unknown) => {
      console.error('获取计划列表失败:', err)
    })
    .finally(() => {
      planLoading.value = false
      refreshTime.value = dayjs()
    })
}

const onEnable = (row: Models.AutoIngestPlan) => {
  enableAutoIngestPlan({ id: row.id })
    .then((res: ApiResponse) => {
      if (res.code === 200) {
        message.success('已启用')
        fetchPlanList()
      }
    })
    .catch((err: unknown) => {
      console.error('启用失败', err)
    })
}

const onDisable = (row: Models.AutoIngestPlan) => {
  disableAutoIngestPlan({ id: row.id })
    .then((res: ApiResponse) => {
      if (res.code === 200) {
        message.success('已停用')
        fetchPlanList()
      }
    })
    .catch((err: unknown) => {
      console.error('停用失败', err)
    })
}

const onDelete = (row: Models.AutoIngestPlan) => {
  deleteAutoIngestPlan({ id: row.id })
    .then((res: ApiResponse) => {
      if (res.code === 200) {
        message.success('删除成功')
        fetchPlanList()
      }
    })
    .catch((err: unknown) => {
      console.error('删除失败', err)
    })
}

// Create Plan
const showCreateModal = ref(false)
const createSubmitting = ref(false)
const createFormRef = ref<FormInst | null>(null)
const createForm = reactive<
  CreateSubscribePlanRequest & {
    refreshStrategy: NonNullable<CreateSubscribePlanRequest['refreshStrategy']>
  }
>({
  name: '',
  parentPath: '',
  upUserId: '',
  cloudToken: undefined,
  onConflict: 'rename',
  autoIngestInterval: 30,
  oneClickAddHistory: false,
  refreshStrategy: {
    enableAutoRefresh: false,
    autoRefreshDays: 7,
    refreshInterval: 30,
    enableDeepRefresh: false,
  },
})

const createRules: FormRules = {
  name: [{ required: true, message: '请输入计划名称', trigger: 'blur' }],
  parentPath: [{ required: true, message: '请输入父目录路径', trigger: 'blur' }],
  upUserId: [{ required: true, message: '请输入上传用户ID', trigger: 'blur' }],
  autoIngestInterval: [
    {
      type: 'number',
      min: AUTO_INGEST_INTERVAL_MIN,
      message: `间隔不能小于${AUTO_INGEST_INTERVAL_MIN}分钟`,
      trigger: 'blur',
    },
  ],
  'refreshStrategy.refreshInterval': [
    {
      type: 'number',
      min: REFRESH_INTERVAL_MIN,
      message: `刷新间隔不能小于${REFRESH_INTERVAL_MIN}分钟`,
      trigger: 'blur',
    },
  ],
  'refreshStrategy.autoRefreshDays': [
    {
      type: 'number',
      min: AUTO_REFRESH_DAYS_MIN,
      message: `持续天数不能小于${AUTO_REFRESH_DAYS_MIN}`,
      trigger: 'blur',
    },
  ],
}

const handleCreatePlan = () => {
  createFormRef.value?.validate((errors) => {
    if (errors) {
      message.error('请检查表单输入')
      return
    }

    createSubmitting.value = true
    createSubscribePlan(createForm)
      .then((res: ApiResponse<{ id: number }>) => {
        if (res.code === 200) {
          message.success('创建成功')
          showCreateModal.value = false
          // 重置部分字段
          createForm.name = ''
          createForm.parentPath = ''
          createForm.upUserId = ''
          createForm.cloudToken = undefined
          createForm.oneClickAddHistory = false
          createForm.onConflict = 'rename'
          createForm.autoIngestInterval = 30
          createForm.refreshStrategy.enableAutoRefresh = false
          createForm.refreshStrategy.autoRefreshDays = 7
          createForm.refreshStrategy.refreshInterval = 30
          createForm.refreshStrategy.enableDeepRefresh = false

          fetchPlanList()
        }
      })
      .catch((err: unknown) => {
        console.error('创建订阅计划失败:', err)
      })
      .finally(() => {
        createSubmitting.value = false
      })
  })
}

// -------- Logs --------
const logLoading = ref(false)
const logTable = ref<Models.AutoIngestLog[]>([])
const logQuery = reactive<{
  planId?: number
  level?: 'info' | 'warn' | 'error'
}>({
  planId: undefined,
  level: undefined,
})
const planOptions = ref<{ label: string; value: number | undefined }[]>([
  { label: '全部计划', value: undefined },
])
const logLevelOptions = [
  { label: '全部级别', value: undefined },
  { label: 'info', value: 'info' },
  { label: 'warn', value: 'warn' },
  { label: 'error', value: 'error' },
]

// 分页（Logs 同步风格）
const logPagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  prefix: ({ itemCount }) => `共 ${itemCount} 条`,
  onChange: (page: number) => {
    logPagination.page = page
    fetchLogList()
  },
  onUpdatePageSize: (ps: number) => {
    logPagination.pageSize = ps
    logPagination.page = 1
    fetchLogList()
  },
})

const handleLogFilter = () => {
  logPagination.page = 1
  fetchLogList()
}
const handleLogReset = () => {
  logQuery.planId = undefined
  logQuery.level = undefined
  logPagination.page = 1
  fetchLogList()
}

const logColumns: DataTableColumns<Models.AutoIngestLog> = [
  { title: 'ID', key: 'id', width: 90, align: 'center' },
  { title: '计划ID', key: 'planId', width: 100, align: 'center' },
  {
    title: '级别',
    key: 'level',
    width: 100,
    align: 'center',
    render: (row) => {
      const type = row.level === 'error' ? 'error' : row.level === 'warn' ? 'warning' : 'success'
      return h(NTag, { type, size: 'small', bordered: true }, { default: () => row.level })
    },
  },
  {
    title: '内容',
    key: 'content',
    minWidth: 360,
    align: 'center',
    render: (row) => h('div', { class: 'log-content' }, row.content),
  },
  {
    title: '时间',
    key: 'createdAt',
    width: 180,
    align: 'center',
    render: (row) => formatDT(row.createdAt),
  },
]

const fetchLogList = () => {
  logLoading.value = true
  getAutoIngestLogList({
    currentPage: logPagination.page || 1,
    pageSize: logPagination.pageSize || 10,
    planId: logQuery.planId || undefined,
    level: logQuery.level || undefined,
  })
    .then((res: ApiResponse<Models.PaginationResponse<Models.AutoIngestLog>>) => {
      if (res.code === 200 && res.data) {
        logTable.value = res.data.data
        logPagination.itemCount = res.data.total
      }
    })
    .catch((err: unknown) => {
      console.error('获取日志失败:', err)
    })
    .finally(() => {
      logLoading.value = false
      refreshTime.value = dayjs()
    })
}

// Utils
const formatDT = (v?: string) => (v ? dayjs(v).format('YYYY-MM-DD HH:mm:ss') : '-')

// Init
onMounted(() => {
  loadCloudTokens()
  fetchPlanList()
  fetchLogList()
})
</script>

<style scoped>
.autoingest-page {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 头部样式参考用户组管理 */
.header {
  margin: 8px 0 4px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
}

.autoingest-table {
  background: white;
  border-radius: 6px;
}

.autoingest-table :deep(.n-data-table-th) {
  text-align: center;
  font-weight: 600;
}

.autoingest-table :deep(.n-data-table-td) {
  text-align: center;
}

.cell-title {
  font-weight: 600;
}

.log-content {
  white-space: pre-wrap;
  word-break: break-all;
}

/* 时间列样式（合并为一列） */
.time-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.time-line {
  display: flex;
  gap: 6px;
  align-items: center;
  font-size: 12px;
}

.time-label {
  color: var(--n-text-color-2);
}

.time-value {
  color: var(--n-text-color);
}
</style>
