package file

import (
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/taskcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

// HandleBatchDelete 后台排队删除处理逻辑
func (h *handler) HandleBatchDelete() taskcontext.HandlerFunc {
	return func(ctx *taskcontext.Context) error {
		req := new(topic.FileBatchDeleteRequest)

		if err := ctx.Unmarshal(req); err != nil {
			h.logger.Error("解析删除任务失败", zap.Error(err))
			return nil
		}

		h.logger.Info("消费者开始处理批量删除", zap.Int("count", len(req.IDs)))

		for _, id := range req.IDs {
			targetFileID := id
			if err := h.deleteMountPointAndFiles(ctx.GetContext(), targetFileID); err != nil {
				h.logger.Error("后台删除挂载点失败", zap.Int64("fid", targetFileID), zap.Error(err))
			}
			time.Sleep(20 * time.Millisecond)
		}

		return nil
	}
}
