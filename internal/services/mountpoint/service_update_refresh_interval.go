package mountpoint

import (
	"errors"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"go.uber.org/zap"
)

func (s *service) UpdateRefreshInterval(ctx context.Context, fileId int64, interval int) error {
	// 验证刷新间隔，最小值为30分钟
	if interval > 0 && interval < 30 {
		return errors.New("刷新间隔最小值为30分钟")
	}

	if err := s.getDB(ctx).Where("file_id = ?", fileId).Update("refresh_interval", interval).Error; err != nil {
		ctx.Error("更新挂载点刷新间隔失败", zap.Error(err), zap.Int64("fileId", fileId), zap.Int("interval", interval))

		return err
	}

	return nil
}
