package file

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
	"go.uber.org/zap"
)

// 为了 实现 hook 功能，添加中间函数操作
func (h *handler) batchDeleteFiles(ctx context.Context, filesToDelete []*models.VirtualFile) (err error) {
	ids := make([]int64, 0, len(filesToDelete))
	for _, file := range filesToDelete {
		ids = append(ids, file.ID)
	}

	if _, err = h.virtualFileService.BatchDelete(ctx, ids); err != nil {
		ctx.Error("批量删除文件 - 服务层删除失败", zap.Int64s("file_ids", ids), zap.Error(err))

		return err
	}

	for _, file := range filesToDelete {
		if file.IsDir {
			if child, childErr := h.virtualFileService.List(ctx, &virtualfile.ListRequest{
				ParentId: &file.ID,
			}); childErr != nil {
				ctx.Error("批量删除文件 - 服务层查询子节点失败", zap.Int64("file_id", file.ID), zap.Error(childErr))

				continue
			} else if len(child) > 0 {
				if err = h.batchDeleteFiles(ctx, child); err != nil {
					ctx.Error("批量删除文件 - 子节点删除失败", zap.Int64("file_id", file.ID), zap.Error(err))

					continue
				}
			}
		}
	}

	return nil
}

func (h *handler) clearMountFiles(ctx context.Context, topId int64) error {
	ctx.Debug("清理挂载文件 - 开始清理", zap.Int64("top_id", topId))

	for {
		files, err := h.virtualFileService.List(ctx, &virtualfile.ListRequest{
			TopId:       &topId,
			CurrentPage: 1,
			PageSize:    1000,
		})
		if err != nil {
			ctx.Error("清理挂载文件 - 服务层查询失败", zap.Int64("top_id", topId), zap.Error(err))

			return err
		}

		if len(files) == 0 {
			break
		}

		fileIds := make([]int64, 0, len(files))
		for _, file := range files {
			fileIds = append(fileIds, file.ID)
		}

		if _, err = h.virtualFileService.BatchDelete(ctx, fileIds); err != nil {
			ctx.Error("批量删除文件 - 服务层删除失败", zap.Int64s("file_ids", fileIds), zap.Error(err))

			continue
		}
	}

	return nil
}

func (h *handler) batchCreateFiles(ctx context.Context, pid int64, filesToCreate []*models.VirtualFile) (err error) {
	_, err = h.virtualFileService.BatchCreate(ctx, pid, filesToCreate)
	if err != nil {
		ctx.Error("批量创建文件 - 服务层创建失败", zap.Int64("pid", pid), zap.Error(err))

		return err
	}

	return nil
}

func (h *handler) batchUpdateFiles(ctx context.Context, filesToUpdate map[int64][]utils.Field) (err error) {
	var texts []string

	for fid, fields := range filesToUpdate {
		if err = h.virtualFileService.Update(ctx, fid, fields); err != nil {
			ctx.Error("批量更新文件 - 服务层更新失败", zap.Int64("file_id", fid), zap.Error(err))

			texts = append(texts, err.Error())
		}
	}

	if len(texts) > 0 {
		return errors.New(strings.Join(texts, "; "))
	}

	return nil
}
