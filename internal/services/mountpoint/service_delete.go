package mountpoint

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"go.uber.org/zap"
)

func (s *service) Delete(ctx context.Context, fileId int64) error {
	if err := s.getDB(ctx).Where("file_id = ?", fileId).Delete(nil).Error; err != nil {
		ctx.Error("删除挂载点失败", zap.Error(err), zap.Int64("fileId", fileId))

		return err
	}

	return nil
}
