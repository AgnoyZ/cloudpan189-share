package mediafile

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"gorm.io/gorm"
)

func (s *service) DeleteStrm(ctx context.Context, fid int64, rootPath string) error {
	file, err := s.QueryStrm(ctx, fid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return err
	}

	// 删除文件
	_ = os.Remove(filepath.Join(rootPath, file.Path))
	// 删除记录
	return s.getDB(ctx).Where("id = ?", file.ID).Delete(new(models.MediaFile)).Error
}
