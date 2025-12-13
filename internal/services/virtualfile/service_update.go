package virtualfile

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UpdateHook func(ctx context.Context, result *gorm.DB, id int64, opts []utils.Field)

func (s *service) Update(ctx context.Context, id int64, opts []utils.Field, hooks ...UpdateHook) error {
	ctx.Debug("更新文件", zap.Int64("file_id", id), zap.Int("field_count", len(opts)))

	updates := make(map[string]interface{})
	for _, opt := range opts {
		updates[opt.Key] = opt.Value
	}

	result := s.withLock(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id).Updates(updates)
	})

	for _, hook := range hooks {
		hook(ctx, result, id, opts)
	}

	return result.Error
}

func (s *service) ModifyAddition(ctx context.Context, id int64, key string, value any) error {
	ctx.Debug("修改文件附加信息", zap.Int64("file_id", id), zap.String("key", key))

	return s.withLock(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id).Update("addition", gorm.Expr("JSON_SET(addition, ?, ?)", "$."+key, value))
	}).Error
}

func (s *service) BatchUpdatePlus(ctx context.Context, values []utils.Field, exps []clause.Expression) error {
	ctx.Debug("批量更新文件", zap.Int("field_count", len(values)), zap.Int("condition_count", len(exps)))

	updates := make(map[string]interface{})
	for _, field := range values {
		updates[field.Key] = field.Value
	}

	query := s.getDB(ctx)
	for _, exp := range exps {
		query = query.Where(exp)
	}

	result := query.Updates(updates)

	return result.Error
}

func (s *service) BatchUpdate(ctx context.Context, filesToUpdate map[int64][]utils.Field) error {
	ctx.Debug("批量更新文件(Map模式)", zap.Int("count", len(filesToUpdate)))

	var err error
	s.withLock(ctx, func(db *gorm.DB) *gorm.DB {
		err = db.Transaction(func(tx *gorm.DB) error {
			for id, fields := range filesToUpdate {
				updates := make(map[string]interface{})
				for _, opt := range fields {
					updates[opt.Key] = opt.Value
				}
				if txErr := tx.Model(&models.VirtualFile{}).Where("id = ?", id).Updates(updates).Error; txErr != nil {
					return txErr
				}
			}
			return nil
		})
		return db
	})

	return err
}