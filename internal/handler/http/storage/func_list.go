package storage

import (
	"github.com/samber/lo"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"

	cloudtokenSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
	filetasklogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	mountpointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
)

type listRequest struct {
	CurrentPage int    `form:"currentPage,omitempty,default=1" binding:"omitempty,min=1" example:"1"` // 当前页码，默认为1
	PageSize    int    `form:"pageSize,omitempty,default=10" binding:"omitempty,min=1" example:"10"`  // 每页大小，默认为10
	Path        string `form:"path" example:"/aaa"`
}

type storageDTO struct {
	ID                    int64                 `json:"id"`
	TaskLogs              []*models.FileTaskLog `json:"taskLogs"`
	TokenName             string                `json:"tokenName"`
	IsInAutoRefreshPeriod bool                  `json:"isInAutoRefreshPeriod"` // 是否在自动刷新时间范围内
	*models.MountPoint
}

type listResponse struct {
	Total       int64         `json:"total" example:"100"`     // 总记录数
	CurrentPage int           `json:"currentPage" example:"1"` // 当前页码
	PageSize    int           `json:"pageSize" example:"10"`   // 每页大小
	Data        []*storageDTO `json:"data"`                    // 列表数据
}

// List 获取存储挂载点列表
// @Summary 获取存储挂载点列表
// @Description 分页获取存储挂载点列表，支持按路径过滤
// @Tags 存储管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param currentPage query int false "当前页码，默认为1" default(1)
// @Param pageSize query int false "每页大小，默认为10" default(10)
// @Param path query string false "路径过滤" example("/aaa")
// @Success 200 {object} httpcontext.Response{data=listResponse} "获取存储挂载点列表成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "查询挂载点失败，code=3019"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/storage/list [get]
func (h *handler) List() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(listRequest)
		if err := ctx.ShouldBindQuery(req); err != nil {
			ctx.AbortWithInvalidParams(err)

			return
		}

		mountReq := &mountpointSvi.ListRequest{
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
			FullPath:    req.Path,
		}

		list, err := h.mountPointService.List(ctx.GetContext(), mountReq)
		if err != nil {
			ctx.Fail(busCodeStorageQueryMountPointError.WithError(err))

			return
		}

		count, err := h.mountPointService.Count(ctx.GetContext(), mountReq)
		if err != nil {
			ctx.Fail(busCodeStorageQueryMountPointError.WithError(err))

			return
		}

		var (
			tokenMap       map[int64]string
			taskLogMapList map[int64][]*models.FileTaskLog
		)

		// 查询令牌名字
		{
			cloudTokenList := make([]int64, 0, len(list))

			for _, item := range list {
				if item.TokenId > 0 {
					cloudTokenList = append(cloudTokenList, item.TokenId)
				}
			}

			cloudTokenList = lo.Uniq(cloudTokenList)

			tokenList, err := h.cloudTokenService.List(ctx.GetContext(), &cloudtokenSvi.ListRequest{
				IdList:     cloudTokenList,
				NoPaginate: true,
			})
			if err != nil {
				ctx.Fail(busCodeStorageQueryCloudTokenError.WithError(err))

				return
			}

			tokenMap = lo.SliceToMap(tokenList, func(item *models.CloudToken) (int64, string) { return item.ID, item.Name })
		}

		// 查询最近的运行任务
		{
			fileIdList := make([]int64, 0, len(list))

			for _, item := range list {
				fileIdList = append(fileIdList, item.FileId)
			}

			fileIdList = lo.Uniq(fileIdList)

			taskLogList, err := h.fileTaskLogService.List(ctx.GetContext(), &filetasklogSvi.ListRequest{
				PageSize:    200,
				CurrentPage: 1,
				FileIdList:  fileIdList,
			})
			if err != nil {
				ctx.Fail(busCodeStorageQueryFileTaskLogError.WithError(err))

				return
			}

			for _, taskLog := range taskLogList {
				if taskLog.FileId == 0 {
					continue
				}

				if taskLogMapList == nil {
					taskLogMapList = make(map[int64][]*models.FileTaskLog)
				}

				taskLogMapList[taskLog.FileId] = append(taskLogMapList[taskLog.FileId], taskLog)
			}
		}

		dtoList := make([]*storageDTO, 0, len(list))

		for _, item := range list {
			tokenName := "令牌未绑定"

			if item.TokenId > 0 {
				if tkName, ok := tokenMap[item.TokenId]; ok {
					tokenName = tkName
				} else {
					tokenName = "令牌不存在"
				}
			}

			taskLogs := make([]*models.FileTaskLog, 0, len(taskLogMapList[item.FileId]))
			if taskLogList, ok := taskLogMapList[item.FileId]; ok {
				taskLogs = taskLogList
			}

			dtoList = append(dtoList, &storageDTO{
				ID:                    item.FileId,
				TaskLogs:              taskLogs,
				TokenName:             tokenName,
				MountPoint:            item,
				IsInAutoRefreshPeriod: item.IsInAutoRefreshPeriod(),
			})
		}

		ctx.Success(&listResponse{
			Total:       count,
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
			Data:        dtoList,
		})
	}
}
