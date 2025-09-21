<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    :class="['cloudpan-logo', variant]"
  >
    <!-- 简约云朵 - 右侧更饱满的形状 -->
    <path
      d="M20 11C21.9 11 23.5 12.6 23.5 14.5S21.9 18 20 18H5.5C3.8 18 2.5 16.7 2.5 15S3.8 12 5.5 12C5.5 9 7.5 6.5 10.5 6.5C13.1 6.5 15.2 8.3 15.6 10.8C16.5 10.5 17.4 10.8 18.2 11.4C18.8 10.9 19.5 10.6 20.2 10.7C20.1 10.8 20 10.9 20 11Z"
      fill="currentColor"
    />

    <!-- 上下浮动的上传标识 -->
    <g>
      <animateTransform
        attributeName="transform"
        attributeType="XML"
        type="translate"
        values="0,0; 0,-1; 0,0; 0,1; 0,0"
        dur="2s"
        repeatCount="indefinite"
      />

      <!-- 上传箭头 -->
      <path
        d="M12 16V11M12 11L10 13M12 11L14 13"
        stroke="white"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </g>
  </svg>
</template>

<script setup lang="ts">
interface Props {
  size?: number | string
  variant?: 'default' | 'collapsed' | 'folder' | 'geometric' | 'server'
}

withDefaults(defineProps<Props>(), {
  size: 24,
  variant: 'default',
})
</script>

<style scoped>
.cloudpan-logo {
  color: var(--n-primary-color);
  flex-shrink: 0;
  transition: color 0.3s ease;
}

.cloudpan-logo.default {
  /* 默认样式 */
}

.cloudpan-logo.collapsed {
  /* 折叠状态样式 */
}

.cloudpan-logo:hover {
  opacity: 0.8;
}
</style>
