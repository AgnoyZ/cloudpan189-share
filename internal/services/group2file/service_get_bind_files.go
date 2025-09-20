package group2file

import (
	"strconv"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"go.uber.org/zap"
)

// GetBindFiles 获取组的所有文件ID
func (s *service) GetBindFiles(ctx context.Context, groupId int64) ([]int64, error) {
	db := s.getDB(ctx)
	groupSubject := s.getGroupSubject(groupId)

	var group2Files []models.Group2File
	if err := db.Where("group_id = ? AND ptype = ?", groupSubject, "p").Find(&group2Files).Error; err != nil {
		ctx.Error("查询用户组文件权限失败", zap.Int64("groupId", groupId), zap.Error(err))
		return nil, err
	}

	var fileIds []int64

	for _, g2f := range group2Files {
		// 解析文件ID，格式为 "fid:123"
		if len(g2f.FileId) > 4 && g2f.FileId[:4] == "fid:" {
			if fileId, err := strconv.ParseInt(g2f.FileId[4:], 10, 64); err == nil {
				fileIds = append(fileIds, fileId)
			}
		}
	}

	ctx.Info("获取用户组绑定文件成功", zap.Int64("groupId", groupId), zap.Int("fileCount", len(fileIds)))

	return fileIds, nil
}
