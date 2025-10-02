<template>
  <img :src="logoSrc" :width="24" :height="24" class="cloudpan-logo" alt="CloudPan Logo" />
</template>

<script setup lang="ts">
// 使用 Vite public 资源路径
const logoSrc = '/logo.png'
</script>

<style scoped>
.cloudpan-logo {
  color: var(--n-primary-color);
  flex-shrink: 0;
  transition: color 0.3s ease;
}

.cloudpan-logo:hover {
  opacity: 0.8;
}
</style>
