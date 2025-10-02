package file

import (
	"path"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	verifySvi "github.com/xxcheng123/cloudpan189-share/internal/services/verify"
	"github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
)

// 为了 实现 hook 功能，添加中间函数操作
func (h *handler) batchDeleteFiles(ctx context.Context, filesToDelete []*models.VirtualFile) (err error) {
	ids := make([]int64, 0, len(filesToDelete))
	for _, file := range filesToDelete {
		ids = append(ids, file.ID)
	}

	if _, err = h.virtualFileService.BatchDelete(ctx, ids, h.deleteStrmIterator); err != nil {
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

		if _, err = h.virtualFileService.BatchDelete(ctx, fileIds, h.deleteStrmIterator); err != nil {
			ctx.Error("批量删除文件 - 服务层删除失败", zap.Int64s("file_ids", fileIds), zap.Error(err))

			continue
		}
	}

	return nil
}

func (h *handler) batchCreateFiles(ctx context.Context, pid int64, filesToCreate []*models.VirtualFile) (err error) {
	_, err = h.virtualFileService.BatchCreate(ctx, pid, filesToCreate, h.createStrmIteratorfunc)
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

func (h *handler) createStrmIteratorfunc(ctx context.Context, result *gorm.DB, files []*models.VirtualFile) {
	if result.Error == nil && shared.MediaConfig != nil && shared.MediaConfig.Enable {
		dirPath, ok := ctx.GetString(consts.CtxKeyFileFullPath)
		if !ok {
			ctx.Error("批量创建文件 - 获取文件路径失败")

			return
		}

		ctx.Debug("批量创建文件 - 创建 strm 文件", zap.Int("file_count", len(files)), zap.String("full_path", dirPath))

		for _, file := range files {
			if file.IsDir {
				continue
			}

			// 获取文件后缀
			extName := path.Ext(file.Name)
			if len(shared.MediaConfig.IncludedSuffixes) > 0 && !slices.Contains(shared.MediaConfig.IncludedSuffixes, extName) {
				ctx.Debug("批量创建文件 - 跳过文件", zap.String("file_name", file.Name))

				continue
			}

			// 重新生成文件名: vidoe.mp4 -> video.strm
			filename := strings.TrimSuffix(file.Name, extName) + ".strm"

			// 生成URL
			values, err := h.verifyService.SignV1(ctx, file.ID, verifySvi.WithV1NoExpire())
			if err != nil {
				ctx.Error("批量创建文件 - 遍历 - 获取文件签名失败", zap.Int64("file_id", file.ID), zap.Error(err))

				continue
			}

			if id, err := h.mediaFileService.WriteStrm(ctx, shared.MediaConfig.GetCar(dirPath, filename), file.ID, shared.JoinDownloadURL(file.ID, values)); err != nil {
				ctx.Error("批量创建文件 - 遍历 - 创建 strm 文件失败", zap.Int64("file_id", file.ID), zap.Error(err))

				continue
			} else {
				ctx.Debug("批量创建文件 - 遍历 - 创建 strm 文件成功", zap.Int64("file_id", file.ID), zap.Int64("media_file_id", id))
			}
		}
	}
}

func (h *handler) deleteStrmIterator(ctx context.Context, result *gorm.DB, idList []int64) {
	if result.Error == nil && shared.MediaConfig != nil && shared.MediaConfig.Enable {
		ctx.Debug("批量删除文件 - 删除 strm 文件", zap.Int("file_count", len(idList)))

		for _, fileId := range idList {
			if err := h.mediaFileService.DeleteStrm(ctx, fileId, shared.MediaConfig.StoragePath); err != nil {
				ctx.Error("批量删除文件 - 遍历 - 删除 strm 文件失败", zap.Int64("file_id", fileId), zap.Error(err))
			}
		}
	}
}
