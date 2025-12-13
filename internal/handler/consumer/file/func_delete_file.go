package file

import (
	"encoding/json"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

// HandleBatchDelete 后台排队删除处理逻辑
func (h *handler) HandleBatchDelete(ctx context.Context, payload []byte) error {
	req := new(topic.FileBatchDeleteRequest)
	if err := json.Unmarshal(payload, req); err != nil {
		// 格式都不对，直接忽略
		return nil
	}

	// 记录一下这批有多少个
	ctx.Info("消费者开始处理批量删除", zap.Int("count", len(req.IDs)))

	for _, id := range req.IDs {
		// 1. 先尝试删除 MountPoint 记录
		if err := h.mountPointService.BatchDelete(ctx, []int64{id}); err != nil {
			// 只有数据库报错才记录 Error，没删掉(rows=0)不算错
			ctx.Warn("后台删除挂载点记录异常", zap.Int64("id", id), zap.Error(err))
		}

		// 2. 尝试删除 VirtualFile (文件本体)
		targetFileID := id
		fileInfo, fileErr := h.virtualFileService.Query(ctx, targetFileID)

		if fileErr != nil {
			// ctx.Debug("文件已不存在，跳过删除", zap.Int64("fid", targetFileID))
			continue
		}

		if fileInfo != nil {
			// 3. 执行真正的递归删除
			// 这里的 batchDeleteFiles 内部会处理子文件
			if delErr := h.batchDeleteFiles(ctx, []*models.VirtualFile{fileInfo}); delErr != nil {
				ctx.Error("后台删除虚拟文件失败", zap.Int64("fid", targetFileID), zap.Error(delErr))
			} else {
				// 成功日志，可以改为 Debug 减少刷屏
				ctx.Info("后台删除完成", zap.Int64("fid", targetFileID))
			}
		}

		// 4. 这里的 Sleep 是为了防止数据库被锁死 (Database Locked)
		// 如果你觉得删得太慢，可以把 50ms 改成 10ms，或者 0
		time.Sleep(20 * time.Millisecond)
	}

	return nil
}