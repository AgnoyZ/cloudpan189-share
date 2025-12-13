package storage

import (
	"encoding/json"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

type batchDeleteRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}

func (h *handler) BatchDelete() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(batchDeleteRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		task := &topic.FileBatchDeleteRequest{IDs: req.IDs}

		body, _ := json.Marshal(task)

		err := h.taskEngine.PushMessage(
			ctx.GetContext().WithValue(consts.CtxKeyInvokeHandlerName, "API批量删除"),
			task.Topic(),
			body, // 传入字节数组
		)

		if err != nil {
			ctx.GetContext().Error("推送删除任务失败", zap.Error(err))
			ctx.AbortWithError(busCodeStorageSendTaskFail, err)
			return
		}

		ctx.GetContext().Info("批量删除请求已加入队列", zap.Int("count", len(req.IDs)))
		ctx.Success("删除任务已提交，后台处理中")
	}
}
