package mountpoint

import (
	"strings"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"go.uber.org/zap"
)

type CreateRequest struct {
	FileId   int64  `json:"fileId"`    // 文件ID
	FullPath string `json:"fullPath" ` // 完整路径
	OsType   string `json:"osType"`    // 操作系统类型
	TokenId  int64  `json:"tokenId"`   // 令牌ID
}

func (s *service) Create(ctx context.Context, req *CreateRequest) (int64, error) {
	// 从完整路径中提取名称 - 使用 split "/" 取最后一个
	parts := strings.Split(req.FullPath, "/")
	name := parts[len(parts)-1]

	if name == "" {
		name = "root"
	}

	mountPoint := &models.MountPoint{
		FileId:   req.FileId,
		Name:     name,
		FullPath: req.FullPath,
		OsType:   req.OsType,
		TokenId:  req.TokenId,
	}

	if err := s.getDB(ctx).Create(mountPoint).Error; err != nil {
		ctx.Error("创建挂载点失败", zap.Error(err), zap.Int64("fileId", req.FileId), zap.String("fullPath", req.FullPath))

		return 0, err
	}

	return mountPoint.ID, nil
}
