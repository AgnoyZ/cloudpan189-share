package storage

import (
	"encoding/json"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

type batchDeleteRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}

// BatchDelete 批量删除存储挂载
func (h *handler) BatchDelete() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(batchDeleteRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		ctx.GetContext().Info("收到批量删除请求", zap.Int64s("ids", req.IDs))

		successCount := 0

		for _, mountPointID := range req.IDs {
			// 1. 先查询挂载点信息 (这一步很重要，因为我们需要知道 FileId)
			mountPointInfo, err := h.mountPointService.Query(ctx.GetContext(), mountPointID)

			// 如果查询失败，说明挂载点记录可能已经不存在了
			// 但我们需要继续尝试删除对应的 VirtualFile，否则前端会一直显示“幽灵文件夹”
			var targetFileID int64
			if err == nil {
				targetFileID = mountPointInfo.FileId
			} else {
				// 如果查不到挂载点信息，我们假设传入的 ID 就是 FileID (兼容前端传 FileID 的情况)
				// 或者是之前遗留的数据
				ctx.GetContext().Warn("查询挂载点失败，尝试直接作为FileID处理", zap.Int64("id", mountPointID), zap.Error(err))
				targetFileID = mountPointID
			}

			// 2. 删除 MountPoint DB 记录 (尝试按 ID 和 FileID 两种方式删，确保删干净)
			// 注意：即使 deleted_rows 为 0 也没关系，我们要确保的是它不在库里
			_ = h.mountPointService.BatchDelete(ctx.GetContext(), []int64{mountPointID})
			if targetFileID != mountPointID {
				_ = h.mountPointService.BatchDelete(ctx.GetContext(), []int64{targetFileID})
			}

			// 3. 删除 VirtualFile 记录
			if targetFileID > 0 {
				fileInfo, fileErr := h.virtualFileService.Query(ctx.GetContext(), targetFileID)
				if fileErr == nil && fileInfo != nil {
					if delErr := h.batchDeleteFiles(ctx.GetContext(), append(make([]*models.VirtualFile, 0), fileInfo)); delErr != nil {
						ctx.GetContext().Error("删除虚拟文件失败", zap.Int64("file_id", targetFileID), zap.Error(delErr))
					} else {
						ctx.GetContext().Info("删除虚拟文件成功", zap.Int64("file_id", targetFileID))
					}
				} else {
					// 查不到文件，可能是已经被删除了，或者 ID 不对
					ctx.GetContext().Warn("未找到对应的虚拟文件，跳过删除", zap.Int64("file_id", targetFileID))
				}

				// 4. 触发后续清理任务 (清理物理文件等)
				_ = h.virtualFileService.ClearUnusedAncestorFolder(ctx.GetContext(), targetFileID)

				taskReq := &topic.FileClearFileRequest{
					FileId: targetFileID,
				}
				body, _ := json.Marshal(taskReq)

				if err := h.taskEngine.PushMessage(
					ctx.GetContext().
						WithValue(consts.CtxKeyInvokeHandlerName, "批量清理执行器"),
					taskReq.Topic(),
					body,
				); err != nil {
					ctx.GetContext().Error("发送清理任务失败", zap.Int64("file_id", targetFileID), zap.Error(err))
				}
			}

			successCount++
			time.Sleep(10 * time.Millisecond)
		}

		ctx.GetContext().Info("批量删除处理完成", zap.Int("req_count", len(req.IDs)), zap.Int("success_count", successCount))
		ctx.Success()
	}
}
