package group2file

import (
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"go.uber.org/zap"
)

// BatchBindFiles 批量绑定文件权限到用户组 先删除 再绑定
func (s *service) BatchBindFiles(ctx context.Context, groupId int64, fileIds []int64) error {
	enforcer := s.svc.GetFileEnforcer()
	groupSubject := s.getGroupSubject(groupId)

	// 1. 从Casbin中删除该用户组的所有策略
	if _, err := enforcer.RemoveFilteredPolicy(0, groupSubject); err != nil {
		ctx.Error("删除Casbin策略失败", zap.String("groupSubject", groupSubject), zap.Error(err))
		return err
	}

	// 2. 批量添加新的文件权限到Casbin
	if len(fileIds) > 0 {
		for _, fileId := range fileIds {
			fileObject := s.getFileObject(fileId)

			// 添加到Casbin，Casbin会自动管理数据库
			if _, err := enforcer.AddPolicy(groupSubject, fileObject, consts.FilePermissionType); err != nil {
				ctx.Error("添加Casbin策略失败", zap.String("groupSubject", groupSubject), zap.String("fileObject", fileObject), zap.Error(err))

				return err
			}
		}
	}

	// 保存Casbin策略到数据库
	if err := enforcer.SavePolicy(); err != nil {
		ctx.Error("保存Casbin策略失败", zap.Error(err))

		return err
	}

	ctx.Info("批量绑定文件权限成功", zap.Int64("groupId", groupId), zap.Int("fileCount", len(fileIds)))

	return nil
}
