import { h } from 'vue'
import { useModal } from 'naive-ui'
import { BatchCreateFromTextModal } from '@/components/storage'

export function useBatchCreateFromTextMount() {
    const modal = useModal()

    const show = (): Promise<{ success: boolean }> => {
        return new Promise((resolve) => {
            const modalInstance = modal.create({
                title: '批量文本导入',
                preset: 'dialog',
                style: {
                    width: '700px',
                },
                content: () =>
                    h(BatchCreateFromTextModal, {
                        // 监听组件发出的 success 事件
                        onSuccess: () => {
                            resolve({ success: true })
                            modalInstance.destroy()
                        },
                        // 监听组件发出的 cancel 事件
                        onCancel: () => {
                            resolve({ success: false })
                            modalInstance.destroy()
                        },
                    }),
                action: () => null, // 隐藏默认的底部按钮，使用组件内部的按钮
                closable: true,
                maskClosable: false, // 禁止点击遮罩关闭，防止误触
            })
        })
    }

    return {
        show,
    }
}
