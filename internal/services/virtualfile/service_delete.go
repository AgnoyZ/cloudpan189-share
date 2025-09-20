package virtualfile

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DeleteHook func(ctx context.Context, result *gorm.DB, id int64)

func (s *service) Delete(ctx context.Context, id int64, hooks ...DeleteHook) error {
	ctx.Debug("删除文件", zap.Int64("file_id", id))

	result := s.withLock(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id).Delete(nil)
	})

	for _, hook := range hooks {
		hook(ctx, result, id)
	}

	return result.Error
}

type BatchDeleteHook func(ctx context.Context, result *gorm.DB, idList []int64)

func (s *service) BatchDelete(ctx context.Context, ids []int64, hooks ...BatchDeleteHook) (deletedIdList []int64, err error) {
	ctx.Debug("批量删除文件", zap.Int64s("file_ids", ids))

	// 如果输入为空，直接返回
	if len(ids) == 0 {
		return []int64{}, nil
	}

	// 查询实际存在的ID
	matchIdList := make([]int64, 0, len(ids))
	if err := s.getDB(ctx).Where("id IN ?", ids).Pluck("id", &matchIdList).Error; err != nil {
		ctx.Error("批量删除文件 - 数据库查询失败", zap.Int64s("file_ids", ids), zap.Error(err))

		return nil, err
	}

	// 如果没有匹配的记录，直接返回
	if len(matchIdList) == 0 {
		ctx.Debug("批量删除文件 - 没有找到匹配的记录", zap.Int64s("file_ids", ids))

		return []int64{}, nil
	}

	// 执行删除操作，使用实际存在的ID列表
	result := s.withLock(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", matchIdList).Delete(nil)
	})

	if result.Error != nil {
		ctx.Error("批量删除文件 - 数据库删除失败", zap.Int64s("file_ids", matchIdList), zap.Error(result.Error))

		return nil, result.Error
	}

	// 执行Hook，传递实际删除的ID列表
	for _, hook := range hooks {
		hook(ctx, result, matchIdList)
	}

	ctx.Debug("批量删除文件完成", zap.Int64s("deleted_ids", matchIdList), zap.Int64("affected_rows", result.RowsAffected))

	return matchIdList, nil
}
