import { api, type ApiResponse } from '@/utils/api'

// 文件搜索相关接口
export interface FileSearchQuery {
  keyword?: string
  pid?: number
  global?: boolean
  pageSize: number
  currentPage: number
}

export interface FileSearchItem extends Models.VirtualFile {
  fullPath: string
}

export interface FileSearchResponse {
  total: number
  currentPage: number
  pageSize: number
  data: FileSearchItem[]
}

// 文件打开相关接口
export interface FileChild extends Models.VirtualFile {
  href: string
}

export interface FileOpenResponse extends Models.VirtualFile {
  href: string
  children?: FileChild[]
  childrenTotal: number
}

// 下载链接相关接口
export interface CreateDownloadUrlRequest {
  fileId: number
}

export interface CreateDownloadUrlResponse {
  downloadUrl: string
}

// 路径编码工具函数
const encodePath = (path: string): string => {
  // 检查是否已经编码过（简单检查是否包含%符号）
  if (path.includes('%')) {
    return path
  }
  return encodeURIComponent(path)
}

// ===== 文件管理接口 =====

// 搜索文件
export const searchFiles = (params: FileSearchQuery): Promise<ApiResponse<FileSearchResponse>> => {
  return api.get('/file/search', { params }).then((res) => res.data)
}

// 打开文件/目录
export const openFile = (fullPath: string): Promise<ApiResponse<FileOpenResponse>> => {
  // 使用安全的路径编码
  const safePath = encodePath(fullPath)
  return api.get(`/file/open/${safePath}`).then((res) => res.data)
}

// 创建下载链接
export const createDownloadUrl = (
  data: CreateDownloadUrlRequest
): Promise<ApiResponse<CreateDownloadUrlResponse>> => {
  return api.post('/file/create_download_url', data).then((res) => res.data)
}
