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

		// 循环处理，因为我们需要获取每个挂载点的详细信息(特别是 FileId)来触发清理任务
		for _, mountPointID := range req.IDs {
			// 1. 先查询挂载点信息，获取关联的 FileId (虚拟文件ID)
			mountPointInfo, err := h.mountPointService.Query(ctx.GetContext(), mountPointID)
			if err != nil {
				continue
			}

			// 2. 调用服务层删除挂载点记录
			if err = h.mountPointService.BatchDelete(ctx.GetContext(), []int64{mountPointID}); err != nil {
				// 修正 Logger 调用：使用 ctx.GetContext().Error
				ctx.GetContext().Error("批量删除 - 删除挂载点DB失败", zap.Int64("id", mountPointID), zap.Error(err))
				continue
			}

			// 3. 清理无用的祖先目录
			_ = h.virtualFileService.ClearUnusedAncestorFolder(ctx.GetContext(), mountPointInfo.FileId)

			// 4. 发送异步清理任务
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

			time.Sleep(10 * time.Millisecond)
		}

		ctx.Success()
	}
}
