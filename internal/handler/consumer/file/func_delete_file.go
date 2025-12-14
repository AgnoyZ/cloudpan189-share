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
			fileInfo, fileErr := h.virtualFileService.Query(ctx.GetContext(), targetFileID)

			if err := h.mountPointService.BatchDelete(ctx.GetContext(), []int64{id}); err != nil {
				h.logger.Debug("后台删除挂载点记录异常(或已删除)", zap.Int64("id", id), zap.Error(err))
			}

			if fileErr != nil || fileInfo == nil {
				h.logger.Warn("虚拟文件不存在或查询失败，跳过清理", zap.Int64("fid", targetFileID), zap.Error(fileErr))
				continue
			}

			if err := h.clearMountFiles(ctx.GetContext(), targetFileID); err != nil {
				h.logger.Error("清理挂载点子文件失败", zap.Int64("fid", targetFileID), zap.Error(err))
			}

			if err := h.virtualFileService.Delete(ctx.GetContext(), targetFileID); err != nil {
				h.logger.Error("删除根虚拟文件失败", zap.Int64("fid", targetFileID), zap.Error(err))
			} else {
				h.logger.Info("后台删除虚拟文件完成", zap.Int64("fid", targetFileID))

				if err := h.virtualFileService.ClearUnusedAncestorFolder(ctx.GetContext(), fileInfo.ParentId); err != nil {
					h.logger.Warn("清理祖先目录失败", zap.Int64("parentId", fileInfo.ParentId), zap.Error(err))
				} else {
					h.logger.Debug("祖先目录清理检查完毕", zap.Int64("parentId", fileInfo.ParentId))
				}
			}
			time.Sleep(20 * time.Millisecond)
		}

		return nil
	}
}
