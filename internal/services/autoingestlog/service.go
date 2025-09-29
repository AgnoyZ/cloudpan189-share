package autoingestlog

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
)

// Service 面向 AutoIngestLog 的日志服务接口（与计划服务分离）
type Service interface {
	// 写日志
	Create(ctx context.Context, planID int64, level string, content string) (int64, error)

	// 查日志
	List(ctx context.Context, req *ListRequest) ([]*models.AutoIngestLog, error)
	Count(ctx context.Context, req *ListRequest) (int64, error)
}
