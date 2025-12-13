import { h } from 'vue'
import { useModal } from 'naive-ui'
import { BatchCreateFromTextModal } from '@/components/storage'
import { useMountPointBind } from '@/composables/useMountPointBind'

export function useBatchCreateFromTextMount() {
    const modal = useModal()
    const mountPointBind = useMountPointBind()

    const show = (): Promise<{ success: boolean }> => {
        return new Promise((resolve) => {
            const modalInstance = modal.create({
                title: '批量文本导入',
                preset: 'dialog',
                style: { width: '700px' },
                content: () =>
                    h(BatchCreateFromTextModal, {
                        // 监听解析成功事件
                        onParsed: async (payload: { items: any[], token: number }) => {
                            // 1. 关闭当前文本输入弹窗
                            modalInstance.destroy()

                            // 2. 转换数据格式以适配 MountPointBindModal
                            // MountPointBindModal 期望的数据包含 name, osType, cloudToken 等
                            const mountItems = payload.items.map(item => ({
                                name: item.name,
                                osType: item.osType, // 后端返回的类型: share_folder 或 person_folder
                                shareCode: item.shareCode,
                                shareAccessCode: item.shareAccessCode,
                                fileId: item.fileId, // 这里包含了 FileID
                                cloudToken: payload.token, // 绑定解析时使用的Token
                                disableSwitchCloudToken: false // 允许用户在下一步修改Token(如果需要)
                            }))

                            // 3. 打开绑定弹窗 (它支持批量设置路径前缀)
                            const result = await mountPointBind.show(mountItems, payload.token)

                            // 4. 返回最终结果
                            if (result && result.length > 0) {
                                resolve({ success: true })
                            } else {
                                resolve({ success: false })
                            }
                        },
                        onCancel: () => {
                            resolve({ success: false })
                            modalInstance.destroy()
                        },
                    }),
                closable: true,
                maskClosable: false,
            })
        })
    }

    return { show }
}
