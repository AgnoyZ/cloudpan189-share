package file

import (
	"path"

	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/ptr"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	virtualfileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
	"gorm.io/gorm"
)

type openRequest struct {
	FullPath string `uri:"fullPath"`
}

const (
	openBaseURL = "/api/file/open"
)

type childDTO struct {
	*models.VirtualFile
	Href string `json:"href"`
}

type openResponse struct {
	*models.VirtualFile
	Href          string      `json:"href"`
	Children      []*childDTO `json:"children,omitempty"`
	ChildrenTotal int64       `json:"childrenTotal"`
}

func (h *handler) Open() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(openRequest)
		if err := ctx.ShouldBindUri(req); err != nil {
			ctx.AbortWithInvalidParams(err)

			return
		}

		// 切分路径查找文件
		paths, err := utils.SplitPath(req.FullPath)
		if err != nil {
			ctx.Fail(busCodeFilePathSplitError.WithError(err))

			return
		}

		var (
			file *models.VirtualFile
		)

		// 根节点
		if len(paths) == 0 {
			file = models.RootFile()
		} else if !utils.CheckIsPath(req.FullPath) {
			ctx.Fail(busCodeFileInvalidPath)

			return
		} else if file, err = h.virtualFileService.QueryByPath(ctx.GetContext(), req.FullPath); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.Fail(busCodeFileNotFound.WithError(err))
			} else {
				ctx.Fail(busCodeFileQueryError.WithError(err))
			}

			return
		}

		var (
			children      []*models.VirtualFile
			childrenCount int64 = 0
		)

		// 查询子节点
		if file.IsDir {
			childReq := &virtualfileSvi.ListRequest{
				ParentId: ptr.Of(file.ID),
			}

			// 目录查询子节点
			if children, err = h.virtualFileService.List(ctx.GetContext(), childReq); err != nil {
				ctx.Fail(busCodeFileQueryError.WithError(err))

				return
			}

			// 子节点数量
			if childrenCount, err = h.virtualFileService.Count(ctx.GetContext(), childReq); err != nil {
				ctx.Fail(busCodeFileQueryError.WithError(err))

				return
			}
		}

		// 转换为 DTO
		childrenDTO := make([]*childDTO, 0, len(children))
		for _, child := range children {
			childrenDTO = append(childrenDTO, &childDTO{
				VirtualFile: child,
				Href:        utils.PathEscape(openBaseURL, path.Join(req.FullPath, child.Name)),
			})
		}

		ctx.Success(&openResponse{
			VirtualFile:   file,
			Href:          utils.PathEscape(openBaseURL, req.FullPath),
			Children:      childrenDTO,
			ChildrenTotal: childrenCount,
		})
	}
}
