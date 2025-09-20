package group2file

import (
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"go.uber.org/zap"
)

// CheckPermission 检查用户组是否有文件访问权限
func (s *service) CheckPermission(ctx context.Context, groupId int64, fileId int64) (bool, error) {
	enforcer := s.svc.GetFileEnforcer()

	groupSubject := s.getGroupSubject(groupId)
	fileObject := s.getFileObject(fileId)

	if groupId == 0 {
		groupSubject = consts.DefaultGroupCheckName
	}

	// 检查是否有读权限
	hasPermission, err := enforcer.Enforce(groupSubject, fileObject, consts.FilePermissionType)
	if err != nil {
		ctx.Error("检查权限失败", zap.Int64("groupId", groupId), zap.Int64("fileId", fileId), zap.Error(err))
		return false, err
	}

	ctx.Debug("权限检查完成", zap.Int64("groupId", groupId), zap.Int64("fileId", fileId), zap.Bool("hasPermission", hasPermission))

	return hasPermission, nil
}
