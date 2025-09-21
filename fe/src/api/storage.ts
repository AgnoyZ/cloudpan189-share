import { api, type ApiResponse } from '@/utils/api'
import type { PaginationResponse, FileTaskLog, MountPoint } from '@/types/models'

// ===== 存储挂载相关接口 =====

// 添加存储挂载请求接口
export interface AddStorageRequest {
  localPath: string // 本地路径
  osType:
    | 'subscribe'
    | 'subscribe_share_folder'
    | 'share_folder'
    | 'person_folder'
    | 'family_folder' // 存储类型
  cloudToken?: number // 云盘令牌ID
  familyId?: string // 家庭云ID
  fileId?: string // 文件ID
  shareAccessCode?: string // 分享访问码
  shareCode?: string // 分享码
  subscribeUser?: string // 订阅用户
}

export interface AddStorageResponse {
  id: number // 存储ID
  path: string // 存储路径
}

// 删除存储挂载请求接口
export interface DeleteStorageRequest {
  id: number // 存储节点ID
}

// 刷新存储挂载请求接口
export interface RefreshStorageRequest {
  id: number // 挂载点ID
  deep?: boolean // 深度刷新
}

// 存储挂载列表查询参数
export interface StorageListQuery {
  currentPage?: number // 当前页码，默认为1
  pageSize?: number // 每页大小，默认为10
  path?: string // 路径过滤
}

// 存储信息接口（扩展挂载点，包含关联数据）
export interface StorageInfo extends MountPoint {
  tokenName?: string // 关联的token名称
  taskLogs?: FileTaskLog[] // 关联的任务日志
}

// ===== 存储管理接口 =====

// 添加存储挂载
export const addStorage = (data: AddStorageRequest): Promise<ApiResponse<AddStorageResponse>> => {
  return api.post('/storage/add', data).then((res) => res.data)
}

// 删除存储挂载
export const deleteStorage = (data: DeleteStorageRequest): Promise<ApiResponse> => {
  return api.post('/storage/delete', data).then((res) => res.data)
}

// 获取存储挂载点列表
export const getStorageList = (
  params?: StorageListQuery
): Promise<ApiResponse<PaginationResponse<StorageInfo>>> => {
  return api.get('/storage/list', { params }).then((res) => res.data)
}

// 刷新存储挂载
export const refreshStorage = (data: RefreshStorageRequest): Promise<ApiResponse> => {
  return api.post('/storage/refresh', data).then((res) => res.data)
}
