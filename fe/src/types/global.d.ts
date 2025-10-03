export namespace StorageType {
  export type Session = Record<string, unknown>
  export type StorageSetting = {
    pathPrefix: string
    selectedToken: number
  }

  export interface Local {
    token: string
    refreshToken: string
    expireTime: number
    user: Models.User
    systemInfo: Models.SystemInfo
    storageSetting: StorageType.StorageSetting
  }
}
