<template>
  <div class="file-list-container">
    <!-- 列表头部 -->
    <div class="list-header">
      <div class="header-item">名称</div>
      <div class="header-item">大小</div>
      <div class="header-item">修改时间</div>
      <div class="header-item">操作</div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <n-spin size="large" />
    </div>

    <!-- 文件列表 -->
    <div v-else class="file-list">
      <div v-for="file in fileList" :key="file.id" class="file-item" @click="handleFileClick(file)">
        <div class="file-info">
          <div class="file-icon-name">
            <n-icon :component="getFileIcon(file.name, file.isDir)" class="file-icon" />
            <span class="file-name">{{ file.name }}</span>
          </div>
          <div class="file-size">
            {{ file.isDir ? '-' : formatFileSize(file.size) }}
          </div>
          <div class="file-date">
            {{ formatDate(file.modifyDate) }}
          </div>
          <div class="file-actions" @click.stop>
            <n-button size="small" type="primary" text @click="handleFileClick(file)">
              {{ file.isDir ? '打开' : '查看' }}
            </n-button>
            <n-button
              v-if="!file.isDir"
              size="small"
              type="error"
              text
              @click="$emit('download', file)"
            >
              下载
            </n-button>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="fileList.length === 0" class="empty-state">
        <n-empty description="此目录为空" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { NButton, NIcon, NEmpty, NSpin } from 'naive-ui'
import {
  FolderOutline,
  DocumentOutline,
  VideocamOutline,
  MusicalNotesOutline,
  ImageOutline,
  ArchiveOutline,
} from '@vicons/ionicons5'
import type { FileChild } from '@/api/file'
import { formatFileSize, formatDate } from '@/utils/format'

// Props
defineProps<{
  fileList: FileChild[]
  loading: boolean
}>()

// Emits
const emit = defineEmits<{
  fileClick: [file: FileChild]
  download: [file: FileChild]
}>()

// 方法
const handleFileClick = (file: FileChild) => {
  emit('fileClick', file)
}

const getFileIcon = (fileName: string, isDir?: boolean) => {
  if (isDir) return FolderOutline

  const ext = fileName.split('.').pop()?.toLowerCase()
  if (!ext) return DocumentOutline

  if (['mp4', 'avi', 'mkv', 'mov', 'wmv', 'flv', 'webm'].includes(ext)) {
    return VideocamOutline
  }

  if (['mp3', 'wav', 'flac', 'aac', 'ogg', 'wma'].includes(ext)) {
    return MusicalNotesOutline
  }

  if (['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'].includes(ext)) {
    return ImageOutline
  }

  if (['zip', 'rar', '7z', 'tar', 'gz', 'bz2'].includes(ext)) {
    return ArchiveOutline
  }

  return DocumentOutline
}
</script>

<style scoped>
.file-list-container {
  background: #fff;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  overflow: hidden;
}

.list-header {
  display: grid;
  grid-template-columns: 1fr 120px 180px 120px;
  gap: 16px;
  padding: 12px 20px;
  background: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
  font-weight: 500;
  color: #495057;
  font-size: 14px;
}

.header-item {
  display: flex;
  align-items: center;
}

.header-item:nth-child(2),
.header-item:nth-child(3),
.header-item:nth-child(4) {
  justify-content: center;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 60px 20px;
}

.file-list {
  min-height: 400px;
}

.file-item {
  border-bottom: 1px solid #f1f3f4;
  cursor: pointer;
  transition: background-color 0.2s;
}

.file-item:hover {
  background-color: #f8f9fa;
}

.file-item:last-child {
  border-bottom: none;
}

.file-info {
  display: grid;
  grid-template-columns: 1fr 120px 180px 120px;
  gap: 16px;
  padding: 12px 20px;
  align-items: center;
}

.file-icon-name {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.file-icon {
  font-size: 20px;
  color: #007bff;
  flex-shrink: 0;
}

.file-name {
  font-size: 14px;
  color: #212529;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-size {
  font-size: 14px;
  color: #6c757d;
  text-align: center;
}

.file-date {
  font-size: 14px;
  color: #6c757d;
  text-align: center;
}

.file-actions {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.empty-state {
  padding: 60px 20px;
  text-align: center;
}

/* 响应式设计 */
@media (width <= 768px) {
  .list-header,
  .file-info {
    grid-template-columns: 1fr 80px 100px;
    gap: 8px;
    padding: 8px 12px;
  }

  .file-actions {
    flex-direction: column;
    gap: 4px;
  }

  .file-date {
    display: none;
  }
}

@media (width <= 480px) {
  .list-header,
  .file-info {
    grid-template-columns: 1fr 60px;
    gap: 8px;
  }

  .file-size {
    display: none;
  }

  .file-name {
    font-size: 13px;
  }
}
</style>
