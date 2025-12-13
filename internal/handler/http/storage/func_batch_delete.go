package storage

import (
	"encoding/json"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
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
。
		for _, mountPointID := range req.IDs {
			// 1. 先查询挂载点信息，获取关联的 FileId (虚拟文件ID)
			mountPointInfo, err := h.mountPointService.Query(ctx.GetContext(), mountPointID)
			if err != nil {
				// 如果查不到，说明可能已经不存在了，跳过
				continue
			}

			// 2. 调用服务层删除挂载点记录 (这里也可以用 BatchDelete 批量删，但循环里单删更方便控制流程)]
			if err = h.mountPointService.BatchDelete(ctx.GetContext(), []int64{mountPointID}); err != nil {
				ctx.Logger.Error("批量删除 - 删除挂载点DB失败", "id", mountPointID, "err", err)
				continue
			}

			// 3. 清理无用的祖先目录 (VirtualFile Service)
			_ = h.virtualFileService.ClearUnusedAncestorFolder(ctx.GetContext(), mountPointInfo.FileId)

			// 4. 发送异步清理任务 (清理 strm 和物理文件)
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
				ctx.Logger.Error("批量删除 - 发送清理任务失败", "id", mountPointID, "err", err)
			}

			// 简单的限流，防止循环过快导致 DB/MQ 压力
			time.Sleep(10 * time.Millisecond)
		}

		ctx.Success()
	}
}
