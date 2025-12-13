<template>
  <div class="batch-text-container">
    <div class="batch-text-content">
      <div class="header-section">
        <n-alert type="info" show-icon title="使用说明" class="mb-4">
          <template #default>
            <div class="usage-guide">
              <p>支持批量导入天翼云盘分享链接或文件夹ID，每行一条数据。</p>
              <p>支持格式：</p>
              <ul class="format-list">
                <li>分享链接：<code>https://cloud.189.cn/t/AbCdEf</code> （自动提取分享码）</li>
                <li>纯分享码：<code>cloud.189.cn/t/AbCdEf</code></li>
                <li>文件夹ID：<code>123456789012345678</code> （纯数字）</li>
              </ul>
              <p>如果包含访问码，请在链接后用空格隔开，或在下方设置默认访问码。</p>
            </div>
          </template>
        </n-alert>
      </div>

      <n-form
          ref="formRef"
          :model="formModel"
          :rules="rules"
          label-placement="left"
          label-width="100px"
          require-mark-placement="right-hanging"
      >
        <!-- 云盘账号选择 -->
        <n-form-item label="云盘账号" path="cloudToken">
          <n-select
              v-model:value="formModel.cloudToken"
              :options="cloudTokenOptions"
              :loading="state.loadingTokens"
              placeholder="请选择要绑定到的云盘账号"
              clearable
          />
        </n-form-item>

        <!-- 默认访问码 -->
        <n-form-item label="默认访问码" path="shareAccessCode">
          <n-input
              v-model:value="formModel.shareAccessCode"
              placeholder="如果资源行中未指定访问码，将使用此默认值"
          >
            <template #prefix>
              <n-icon :component="LockClosedOutline" />
            </template>
          </n-input>
        </n-form-item>

        <!-- 文本内容 -->
        <n-form-item label="资源列表" path="content">
          <n-input
              v-model:value="formModel.content"
              type="textarea"
              placeholder="请输入资源链接或ID，一行一个&#10;示例：&#10;cloud.189.cn/t/code 1234&#10;88889999"
              :autosize="{ minRows: 8, maxRows: 15 }"
              class="resource-textarea"
          />
        </n-form-item>

        <!-- 自动刷新配置 -->
        <n-divider title-placement="left" dashed>
          <n-text depth="3" style="font-size: 12px">刷新配置</n-text>
        </n-divider>

        <div class="refresh-config-grid">
          <n-form-item label="自动刷新" path="enableAutoRefresh">
            <n-switch v-model:value="formModel.enableAutoRefresh" />
          </n-form-item>

          <template v-if="formModel.enableAutoRefresh">
            <n-form-item label="刷新间隔" path="refreshInterval">
              <n-input-number
                  v-model:value="formModel.refreshInterval"
                  :min="30"
                  :step="30"
                  style="width: 100%"
              >
                <template #suffix>秒</template>
              </n-input-number>
            </n-form-item>
          </template>
        </div>
      </n-form>
    </div>

    <!-- 底部操作栏 -->
    <div class="modal-actions">
      <n-button @click="handleCancel">取消</n-button>
      <n-button
          type="primary"
          :loading="state.submitting"
          @click="handleConfirm"
      >
        开始导入
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted } from 'vue'
import {
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NButton,
  NSwitch,
  NInputNumber,
  NAlert,
  NDivider,
  NIcon,
  NText,
  useMessage,
  type FormInst,
  type FormRules
} from 'naive-ui'
import { LockClosedOutline } from '@vicons/ionicons5'
import { getCloudTokenList } from '@/api/cloudtoken'
import { batchCreateStorageFromText, type BatchCreateTextRequest } from '@/api/storage'

// 定义事件
interface Emits {
  (e: 'success'): void
  (e: 'cancel'): void
}
const emit = defineEmits<Emits>()

// 消息提示
const message = useMessage()

// 表单引用
const formRef = ref<FormInst | null>(null)

// 状态管理
const state = reactive({
  loadingTokens: false,
  submitting: false,
  cloudTokens: [] as Models.CloudToken[]
})

// 表单数据
const formModel = reactive({
  cloudToken: null as number | null,
  content: '',
  shareAccessCode: '',
  enableAutoRefresh: false,
  refreshInterval: 3600
})

// 表单验证规则
const rules: FormRules = {
  cloudToken: [
    {
      type: 'number',
      required: true,
      message: '请选择云盘账号',
      trigger: ['blur', 'change']
    }
  ],
  content: [
    {
      required: true,
      message: '请输入资源列表内容',
      trigger: 'blur'
    }
  ],
  refreshInterval: [
    {
      type: 'number',
      min: 30,
      message: '刷新间隔最小为30秒',
      trigger: 'blur'
    }
  ]
}

// 计算云盘选项
const cloudTokenOptions = computed(() => {
  return state.cloudTokens.map((token) => ({
    label: token.name,
    value: token.id
  }))
})

// 获取云盘列表
const fetchCloudTokens = async () => {
  state.loadingTokens = true
  try {
    const res = await getCloudTokenList({ noPaginate: true })
    if (res.code === 0 || res.code === 200) {
      const rawData = res.data
      let list: Models.CloudToken[] = []

      if (Array.isArray(rawData)) {
        list = rawData
      } else if (rawData && typeof rawData === 'object') {
        // 使用类型断言或可选链
        list = (rawData as any).data || []
      }

      state.cloudTokens = list
      // 如果只有一个账号，默认选中
      if (state.cloudTokens.length === 1) {
        formModel.cloudToken = state.cloudTokens[0].id
      }
    }
  } catch (error) {
    message.error('获取云盘账号列表失败')
  } finally {
    state.loadingTokens = false
  }
}

// 提交处理
const handleConfirm = () => {
  formRef.value?.validate((errors) => {
    if (errors) {
      return
    }

    if (!formModel.cloudToken) return

    state.submitting = true

    const reqData: BatchCreateTextRequest = {
      content: formModel.content,
      cloudToken: formModel.cloudToken,
      shareAccessCode: formModel.shareAccessCode,
      enableAutoRefresh: formModel.enableAutoRefresh,
      refreshInterval: formModel.enableAutoRefresh ? formModel.refreshInterval : undefined
    }

    batchCreateStorageFromText(reqData)
        .then((res) => {
          if (res.code === 200 && res.data) {
            const { total, success, failed } = res.data
            if (failed === 0) {
              message.success(`导入成功！共导入 ${success} 个资源`)
              emit('success')
            } else {
              message.warning(`导入完成：成功 ${success} 个，失败 ${failed} 个（共 ${total} 个）`)
              emit('success')
            }
          } else {
            message.error(res.msg || '导入失败')
          }
        })
        .catch((err) => {
          message.error(err.message || '网络请求异常')
        })
        .finally(() => {
          state.submitting = false
        })
  })
}

const handleCancel = () => {
  emit('cancel')
}

// 初始化
onMounted(() => {
  fetchCloudTokens()
})
</script>

<style scoped>
.batch-text-container {
  width: 100%;
}

.batch-text-content {
  padding: 0 4px;
}

.header-section {
  margin-bottom: 20px;
}

.usage-guide {
  font-size: 13px;
  line-height: 1.6;
}

.usage-guide p {
  margin: 4px 0;
}

.format-list {
  margin: 4px 0 8px 18px;
  color: var(--n-text-color-3);
}

.resource-textarea {
  font-family: monospace;
}

.refresh-config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.mb-4 {
  margin-bottom: 16px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  padding-top: 20px;
  margin-top: 10px;
  border-top: 1px solid var(--n-border-color);
}

@media (max-width: 600px) {
  .refresh-config-grid {
    grid-template-columns: 1fr;
  }
}
</style>
