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
			// 1. 先查询挂载点信息
			mountPointInfo, err := h.mountPointService.Query(ctx.GetContext(), mountPointID)
			if err != nil {
				// 打印错误日志：看看是查不到(RecordNotFound)还是其他错误
				ctx.GetContext().Warn("批量删除 - 查询挂载点失败(可能已删除)", zap.Int64("id", mountPointID), zap.Error(err))
				continue
			}

			// 2. 删除 DB 记录
			if err = h.mountPointService.BatchDelete(ctx.GetContext(), []int64{mountPointID}); err != nil {
				ctx.GetContext().Error("批量删除 - 删除DB记录失败", zap.Int64("id", mountPointID), zap.Error(err))
				continue
			}

			successCount++

			// 3. 触发后续清理逻辑
			if mountPointInfo.FileId > 0 {
				_ = h.virtualFileService.ClearUnusedAncestorFolder(ctx.GetContext(), mountPointInfo.FileId)

				taskReq := &topic.FileClearFileRequest{
					FileId: mountPointInfo.FileId,
				}
				body, _ := json.Marshal(taskReq)

				if err := h.taskEngine.PushMessage(
					ctx.GetContext().
						WithValue(consts.CtxKeyFullPath, mountPointInfo.FullPath).
						WithValue(consts.CtxKeyInvokeHandlerName, "批量清理执行器"),
					taskReq.Topic(),
					body,
				); err != nil {
					ctx.GetContext().Error("批量删除 - 发送清理任务失败", zap.Int64("id", mountPointID), zap.Error(err))
				}
			} else {
				ctx.GetContext().Warn("批量删除 - 挂载点关联的 FileId 为 0", zap.Int64("id", mountPointID))
			}

			time.Sleep(10 * time.Millisecond)
		}

		ctx.GetContext().Info("批量删除完成", zap.Int("req_count", len(req.IDs)), zap.Int("success_count", successCount))
		ctx.Success()
	}
}
