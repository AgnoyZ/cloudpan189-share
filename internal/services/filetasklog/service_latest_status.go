package filetasklog

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"go.uber.org/zap"
)

type latestTaskStatusRow struct {
	FileId int64  `gorm:"column:file_id"`
	Status string `gorm:"column:status"`
}

func (s *service) LatestStatusByFileIDs(ctx context.Context, typ string, fileIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(fileIDs))
	if len(fileIDs) == 0 {
		return result, nil
	}

	latestIDSubQuery := s.getDB(ctx).
		Select("MAX(id) AS id").
		Where("type = ?", typ).
		Where("file_id IN ?", fileIDs).
		Group("file_id")

	rows := make([]*latestTaskStatusRow, 0, len(fileIDs))
	if err := s.getDB(ctx).
		Model(new(models.FileTaskLog)).
		Select("file_id, status").
		Where("id IN (?)", latestIDSubQuery).
		Find(&rows).Error; err != nil {
		ctx.Error("查询文件最新任务状态失败", zap.Error(err), zap.Int("file_count", len(fileIDs)))
		return nil, err
	}

	for _, row := range rows {
		result[row.FileId] = row.Status
	}

	return result, nil
}
