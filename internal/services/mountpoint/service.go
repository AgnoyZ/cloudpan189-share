package mountpoint

import (
	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, req *CreateRequest) (int64, error)
	Query(ctx context.Context, fileId int64) (*models.MountPoint, error)
	QueryByPath(ctx context.Context, fullPath string) (*models.MountPoint, error)
	List(ctx context.Context, req *ListRequest) ([]*models.MountPoint, error)
	Count(ctx context.Context, req *ListRequest) (int64, error)
	Delete(ctx context.Context, fileId int64) error
	EnableAutoRefresh(ctx context.Context, fileId int64, enable bool) error
	UpdateRefreshInterval(ctx context.Context, fileId int64, interval int) error
}

type service struct {
	svc bootstrap.ServiceContext
}

func NewService(svc bootstrap.ServiceContext) Service {
	return &service{
		svc: svc,
	}
}

func (s *service) getDB(ctx context.Context) *gorm.DB {
	return s.svc.GetDB(ctx).Model(new(models.MountPoint))
}
