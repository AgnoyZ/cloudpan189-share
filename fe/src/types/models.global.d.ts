declare namespace Models {
  // 用户模型
  interface User {
    id: number
    username: string
    status: number
    isAdmin: boolean
    groupId: number
    version: number
    createdAt: string
    updatedAt: string
  }

  // 用户信息（包含用户组名称）
  interface UserInfo extends User {
    groupName?: string
  }

  // 云盘令牌模型
  interface CloudToken {
    id: number
    name: string
    username: string
    accessToken: string
    expiresIn: number
    loginType: number // 1: 扫码登录 2: 密码登录
    status: number // 状态 1:正常 2: 登录失败
    addition: Record<string, unknown> // 附属参数
    createdAt: string
    updatedAt: string
  }

  // 用户组模型
  interface UserGroup {
    id: number
    name: string
    createdAt: string
    updatedAt: string
    userCount?: number // 该用户组下的用户数量
  }

  // 虚拟文件模型
  interface VirtualFile {
    id: number
    cloudId: string
    parentId: number
    topId: number
    isTop: boolean
    isDir: boolean
    name: string
    size: number
    hash: string
    osType: string
    addition: Record<string, unknown>
    rev: string
    createDate: string
    modifyDate: string
    createdAt: string
    updatedAt: string
  }

  // 文件任务日志模型（对应后端 FileTaskLog）
  interface FileTaskLog {
    id: number
    title: string
    type: string
    desc: string
    beginAt: string
    endAt: string | null
    status: string
    result: string
    errorMsg: string
    addition: Record<string, unknown>
    duration: number
    fileId: number
    userId: number
    completed: number
    total: number
    createdAt: string
    updatedAt: string
  }

  // 挂载点模型（对应后端 MountPoint）
  interface MountPoint {
    id: number
    fileId: number
    osType: string
    tokenId: number
    name: string
    fullPath: string
    enableAutoRefresh: boolean
    refreshInterval: number
    enableDeepRefresh: boolean
    autoRefreshBeginAt: string
    autoRefreshDays: number
    isInAutoRefreshPeriod: boolean // 是否在自动刷新时间范围内
    lastState: string
    createdAt: string
    updatedAt: string
  }

  // 分页响应基础结构
  interface PaginationResponse<T> {
    currentPage: number
    pageSize: number
    total: number
    data: T[]
  }

  interface SystemInfo {
    baseURL: string
    enableAuth: boolean
    initialized: boolean
    runTime: number // 运行时间 单位 s
    runTimeHuman: string // 运行时间 格式 例如：1年2月3天4小时5分6秒
    title: string
  }
}
