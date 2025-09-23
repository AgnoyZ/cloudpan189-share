declare namespace StorageType {
  type Session = Record<string, unknown>

  interface Local {
    token: string
    refreshToken: string
    expireTime: number
    user: Models.User
  }
}
