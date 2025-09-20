package group2file

import (
	"fmt"

	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"gorm.io/gorm"
)

type Service interface {
	// BatchBindFiles 批量绑定文件权限到用户组 先删除 再绑定
	BatchBindFiles(ctx context.Context, groupId int64, fileIds []int64) error
	// GetBindFiles 获取组的所有文件ID
	GetBindFiles(ctx context.Context, groupId int64) ([]int64, error)
	// CheckPermission 检查用户组是否有文件访问权限
	CheckPermission(ctx context.Context, groupId int64, fileId int64) (bool, error)
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
	return s.svc.GetDB(ctx)
}

// getGroupSubject 获取用户组标识符
func (s *service) getGroupSubject(groupId int64) string {
	return fmt.Sprintf("gid:%d", groupId)
}

// getFileObject 获取文件对象标识符
func (s *service) getFileObject(fileId int64) string {
	return fmt.Sprintf("fid:%d", fileId)
}
