package file

import (
	"path"
	"slices"
	"strings"
	"time"

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

// batchDeleteFiles 递归删除文件
func (h *handler) batchDeleteFiles(ctx context.Context, filesToDelete []*models.VirtualFile) (err error) {
	if len(filesToDelete) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(filesToDelete))
	for _, file := range filesToDelete {
		ids = append(ids, file.ID)
	}

	// 执行当前层级的删除
	if _, err = h.virtualFileService.BatchDelete(ctx, ids, h.deleteStrmIterator); err != nil {
		ctx.Error("批量删除文件 - 服务层删除失败", zap.Int64s("file_ids", ids), zap.Error(err))

		return err
	}

	// 性能优化：删除一批文件后，短暂休眠，释放 DB 锁给前台 Web 请求
	time.Sleep(10 * time.Millisecond)

	for _, file := range filesToDelete {
		if file.IsDir {
			// 更新上下文中的路径，以便子文件的 strm 删除能找到正确路径
			currentPath, _ := ctx.GetString(consts.CtxKeyFileFullPath)
			subCtx := ctx.WithValue(consts.CtxKeyFileFullPath, path.Join(currentPath, file.Name))

			if child, childErr := h.virtualFileService.List(subCtx, &virtualfile.ListRequest{
				ParentId: &file.ID,
			}); childErr != nil {
				// 忽略记录不存在的错误，可能已经被并发删除了
				if !errors.Is(childErr, gorm.ErrRecordNotFound) {
					ctx.Error("批量删除文件 - 服务层查询子节点失败", zap.Int64("file_id", file.ID), zap.Error(childErr))
				}
				continue
			} else if len(child) > 0 {
				if err = h.batchDeleteFiles(subCtx, child); err != nil {
					ctx.Error("批量删除文件 - 子节点删除失败", zap.Int64("file_id", file.ID), zap.Error(err))

					continue
				}
			}
		}
	}

	return nil
}

// clearMountFiles 清理挂载点下的所有文件
func (h *handler) clearMountFiles(ctx context.Context, topId int64) error {
	ctx.Debug("清理挂载文件 - 开始清理", zap.Int64("top_id", topId))

	for {
		files, err := h.virtualFileService.List(ctx, &virtualfile.ListRequest{
			TopId:       &topId,
			CurrentPage: 1,
			PageSize:    500,
		})
		if err != nil {
			ctx.Error("清理挂载文件 - 服务层查询失败", zap.Int64("top_id", topId), zap.Error(err))
			return err
		}

		if len(files) == 0 {
			break
		}

		// 过滤掉挂载点自己（防止死循环，通常 TopId=ID 时 List 会查出来）
		filesToDelete := make([]int64, 0, len(files))
		for _, file := range files {
			if file.ID != topId {
				filesToDelete = append(filesToDelete, file.ID)
			}
		}

		if len(filesToDelete) == 0 {
			break
		}

		if _, err = h.virtualFileService.BatchDelete(ctx, filesToDelete, h.deleteStrmIterator); err != nil {
			ctx.Error("批量删除文件 - 服务层删除失败", zap.Int64s("file_ids", filesToDelete), zap.Error(err))
			// 如果删除失败，避免死循环
			return err
		}

		// 性能优化：每批次处理完，强制休眠，让出 CPU 和 DB 锁
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

func (h *handler) batchCreateFiles(ctx context.Context, pid int64, filesToCreate []*models.VirtualFile) (err error) {
	_, err = h.virtualFileService.BatchCreate(ctx, pid, filesToCreate, h.createStrmIteratorfunc)
	if err != nil {
		ctx.Error("批量创建文件 - 服务层创建失败", zap.Int64("pid", pid), zap.Error(err))

		return err
	}
	// 性能优化：创建操作后也休眠一下
	time.Sleep(10 * time.Millisecond)
	return nil
}

func (h *handler) batchUpdateFiles(ctx context.Context, filesToUpdate map[int64][]utils.Field) (err error) {
	if err = h.virtualFileService.BatchUpdate(ctx, filesToUpdate); err != nil {
		ctx.Error("批量更新文件 - 服务层更新失败", zap.Error(err))
		return err
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

func (h *handler) deleteStrmIterator(ctx context.Context, result *gorm.DB, files []*models.VirtualFile) {
	if result.Error == nil && shared.MediaConfig != nil && shared.MediaConfig.Enable {
		ctx.Debug("批量删除文件 - 删除 strm 文件", zap.Int("file_count", len(files)))

		// 尝试从 Context 获取当前目录的完整路径（这是物理路径的前缀/虚拟路径）
		// 注意：这里的 dirPath 是虚拟路径，我们需要结合 StoragePath 转换为物理路径
		dirPath, hasPath := ctx.GetString(consts.CtxKeyFileFullPath)

		for _, file := range files {
			if file.IsDir {
				continue
			}

			// 逻辑 1: 如果有 fileId，尝试清理数据库记录（如果 media_file 表存在）
			_ = h.mediaFileService.DeleteStrm(ctx, file.ID, shared.MediaConfig.StoragePath)

			// 逻辑 2: 物理删除 strm 文件
			if hasPath {
				extName := path.Ext(file.Name)
				if len(shared.MediaConfig.IncludedSuffixes) > 0 && !slices.Contains(shared.MediaConfig.IncludedSuffixes, extName) {
					continue
				}

				// 构造 strm 文件名
				strmName := strings.TrimSuffix(file.Name, extName) + ".strm"

				fullPhysicalPath := path.Join(shared.MediaConfig.StoragePath, dirPath, strmName)

				if err := h.mediaFileService.DeleteStrmByFullPath(ctx, fullPhysicalPath); err != nil {
					ctx.Warn("删除 strm 物理文件失败", zap.String("path", fullPhysicalPath), zap.Error(err))
				} else {
					ctx.Debug("删除 strm 物理文件成功", zap.String("path", fullPhysicalPath))
				}
			}
		}
	}
}