package file

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/taskcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

func (h *handler) ClearFile() taskcontext.HandlerFunc {
	return func(ctx *taskcontext.Context) error {
		req := new(topic.FileClearFileRequest)

		if err := ctx.Unmarshal(req); err != nil {
			return err
		}

		var (
			logger = ctx.GetContext().Logger
		)

		vf, err := h.virtualFileService.Query(ctx.GetContext(), req.FileId)
		if err != nil {
			logger.Error("查询文件失败", zap.Error(err), zap.Int64("file_id", req.FileId))

			return err
		}

		logger.Debug("开始清理文件", zap.Int64("file_id", vf.ID))

		return h.clearMountFiles(ctx.GetContext(), vf.ID)
	}
}
