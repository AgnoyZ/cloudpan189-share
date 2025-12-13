package file

import (
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/taskcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

// HandleBatchDelete 后台排队删除处理逻辑
func (h *handler) HandleBatchDelete() taskcontext.HandlerFunc {
	return func(ctx taskcontext.Context) error {
		req := new(topic.FileBatchDeleteRequest)

		if err := ctx.Bind(req); err != nil {
			h.logger.Error("解析删除任务失败", zap.Error(err))
			return nil
		}

		h.logger.Info("消费者开始处理批量删除", zap.Int("count", len(req.IDs)))

		for _, id := range req.IDs {
			// 删除挂载点
			if err := h.mountPointService.BatchDelete(ctx.GetContext(), []int64{id}); err != nil {
				h.logger.Warn("后台删除挂载点记录异常", zap.Int64("id", id), zap.Error(err))
			}

			// 删除虚拟文件
			targetFileID := id
			fileInfo, fileErr := h.virtualFileService.Query(ctx.GetContext(), targetFileID)

			if fileErr != nil {
				continue
			}

			if fileInfo != nil {
				if delErr := h.batchDeleteFiles(ctx.GetContext(), []*models.VirtualFile{fileInfo}); delErr != nil {
					h.logger.Error("后台删除虚拟文件失败", zap.Int64("fid", targetFileID), zap.Error(delErr))
				} else {
					h.logger.Info("后台删除完成", zap.Int64("fid", targetFileID))
				}
			}
			time.Sleep(20 * time.Millisecond)
		}

		return nil
	}
}