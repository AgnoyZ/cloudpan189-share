<template>
  <div class="login-container">
    <!-- 背景装饰元素 -->
    <div class="bg-decoration">
      <div class="floating-circle circle-1"></div>
      <div class="floating-circle circle-2"></div>
      <div class="floating-circle circle-3"></div>
      <div class="floating-circle circle-4"></div>
      <div class="floating-circle circle-5"></div>
      <div class="wave wave-1"></div>
      <div class="wave wave-2"></div>
    </div>

    <div class="login-card">
      <div class="login-header">
        <h1>{{ title || '云盘分享系统' }}</h1>
        <p>请登录您的账户</p>
      </div>

      <n-form ref="formRef" :model="formData" :rules="rules" size="large" :show-label="false">
        <n-form-item path="username">
          <n-input
            v-model:value="formData.username"
            placeholder="用户名"
            :input-props="{ autocomplete: 'username' }"
          >
            <template #prefix>
              <n-icon :component="PersonOutline" />
            </template>
          </n-input>
        </n-form-item>

        <n-form-item path="password">
          <n-input
            v-model:value="formData.password"
            type="password"
            placeholder="密码"
            show-password-on="mousedown"
            :input-props="{ autocomplete: 'current-password' }"
            @keydown.enter="handleLogin"
          >
            <template #prefix>
              <n-icon :component="LockClosedOutline" />
            </template>
          </n-input>
        </n-form-item>

        <n-form-item>
          <n-button
            type="primary"
            size="large"
            :loading="loading"
            :block="true"
            @click="handleLogin"
          >
            登录
          </n-button>
        </n-form-item>
      </n-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { PersonOutline, LockClosedOutline } from '@vicons/ionicons5'
import { type LoginRequest } from '@/api/auth'
import { useSystemStore } from '@/stores/system'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const message = useMessage()
const systemStore = useSystemStore()
const userStore = useUserStore()

const title = computed(() => systemStore.systemInfo.title)

const formRef = ref()
const loading = ref(false)

const formData = reactive<LoginRequest>({
  username: '',
  password: '',
})

const rules = {
  username: [
    {
      required: true,
      message: '请输入用户名',
      trigger: ['input', 'blur'],
    },
  ],
  password: [
    {
      required: true,
      message: '请输入密码',
      trigger: ['input', 'blur'],
    },
  ],
}

const handleLogin = () => {
  formRef.value
    ?.validate()
    .then(() => {
      loading.value = true
      return userStore.userLogin(formData)
    })
    .then(() => {
      message.success('登录成功')
      router.push('/@dashboard')
    })
    .catch((error: unknown) => {
      console.error('登录失败:', error)
      message.error('登录失败，请检查用户名和密码')
    })
    .finally(() => {
      loading.value = false
    })
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #74b9ff 0%, #0984e3 50%, #6c5ce7 100%);
  padding: 20px;
  position: relative;
  overflow: hidden;
}

/* 背景装饰元素 */
.bg-decoration {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

/* 浮动圆形装饰 */
.floating-circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  animation: float 6s ease-in-out infinite;
}

.circle-1 {
  width: 80px;
  height: 80px;
  top: 10%;
  left: 10%;
  animation-delay: 0s;
}

.circle-2 {
  width: 120px;
  height: 120px;
  top: 20%;
  right: 15%;
  animation-delay: 1s;
}

.circle-3 {
  width: 60px;
  height: 60px;
  bottom: 30%;
  left: 20%;
  animation-delay: 2s;
}

.circle-4 {
  width: 100px;
  height: 100px;
  bottom: 15%;
  right: 10%;
  animation-delay: 3s;
}

.circle-5 {
  width: 40px;
  height: 40px;
  top: 50%;
  left: 5%;
  animation-delay: 4s;
}

/* 波浪装饰 */
.wave {
  position: absolute;
  width: 200%;
  height: 200px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 50%;
  animation: wave 8s ease-in-out infinite;
}

.wave-1 {
  top: -100px;
  left: -50%;
  animation-delay: 0s;
}

.wave-2 {
  bottom: -100px;
  right: -50%;
  animation-delay: 4s;
}

/* 动画效果 */
@keyframes float {
  0%,
  100% {
    transform: translateY(0px) rotate(0deg);
  }
  50% {
    transform: translateY(-20px) rotate(180deg);
  }
}

@keyframes wave {
  0%,
  100% {
    transform: scale(1) rotate(0deg);
    opacity: 0.3;
  }
  50% {
    transform: scale(1.1) rotate(180deg);
    opacity: 0.1;
  }
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.2);
  position: relative;
  z-index: 2;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  font-size: 28px;
  font-weight: 600;
  color: #333;
  margin: 0 0 8px 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.login-header p {
  font-size: 16px;
  color: #666;
  margin: 0;
}

.n-form-item {
  margin-bottom: 20px;
}

.n-form-item:last-child {
  margin-bottom: 0;
  margin-top: 32px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .floating-circle {
    display: none;
  }

  .wave {
    display: none;
  }

  .login-card {
    margin: 20px;
    padding: 30px;
  }
}
</style>
