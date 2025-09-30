package autoingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/taskcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/datatypes"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/types/autoingest"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
	"gorm.io/gorm"

	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
	storagefacadeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/storagefacade"
)

func (h *handler) RefreshSubscribe() taskcontext.HandlerFunc {
	return func(ctx *taskcontext.Context) error {
		req := new(topic.AutoIngestRefreshSubscribeRequest)

		if err := ctx.Unmarshal(req); err != nil {
			return err
		}

		var (
			pageSize   = 30
			pageNum    = 1
			hasTop     = true
			shouldNext = true

			addCount    int64 = 0
			failedCount int64 = 0
			nextOffset        = req.Offset
		)

		logger := ctx.GetContext().Logger

		defer func() {
			// 执行完了 更新 数据
			_ = h.autoIngestPlanService.UpdateOffset(ctx.GetContext(), req.PlanId, nextOffset)
			_ = h.autoIngestPlanService.IncrAddCount(ctx.GetContext(), req.PlanId, addCount)
			_ = h.autoIngestPlanService.IncrFailedCount(ctx.GetContext(), req.PlanId, failedCount)
		}()

		for shouldNext {
			list, _, err := h.cloudbridgeService.GetSubscribeUserShareResource(ctx.GetContext(), req.UpUserId, func(opt *cloudbridgeSvi.SubscribeUserShareResourceOption) {
				opt.PageNum = pageNum
				opt.PageSize = pageSize
			})
			if err != nil {
				logger.Error("获取订阅号内容时失败了~", zap.String("up_user_id", req.UpUserId), zap.Int("page_num", pageNum), zap.Int("page_size", pageSize))

				return err
			}

			if len(list) == 0 {
				break
			}

			pageNum++

			for _, item := range list {
				if item.IsTop != 1 {
					hasTop = false
				}

				itemOffset := item.ShareTime.Unix()
				if itemOffset > nextOffset {
					nextOffset = itemOffset
				}

				if itemOffset < req.Offset {
					if !hasTop {
						shouldNext = false
					}

					continue
				}

				logger.Debug("发现新的待入库文件", zap.String("name", item.Name))

				// 检查这个文件存不存在先
				fullPath := path.Join(req.ParentPath, item.Name)
				if _, err := h.virtualFileService.QueryByPath(ctx.GetContext(), fullPath); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					logger.Error("查询虚拟文件路径失败 跳过本次自动入库", zap.String("path", fullPath), zap.Error(err))

					continue
				} else if err == nil && req.OnConflict == autoingest.OnConflictRename {
					fullPath = path.Join(req.ParentPath, fmt.Sprintf("%s_%d", item.Name, time.Now().Unix()))
				}

				id, err := h.storageFacadeService.CreateStorage(ctx.GetContext(),
					&storagefacadeSvi.CreateStorageRequest{
						LocalPath:  fullPath,
						OsType:     models.OsTypeSubscribeShareFolder,
						CloudToken: req.CloudToken,
						FileId:     item.ID,
						Addition: datatypes.JSONMap{
							consts.FileAdditionKeyUpUserId: req.UpUserId,
							consts.FileAdditionKeyShareId:  item.ShareId,
							consts.FileAdditionKeyIsFolder: item.IsFolder,
						},
					},
				)
				if err != nil {
					logger.Error("入库失败", zap.Error(err), zap.String("path", fullPath))

					continue
				}

				taskReq := &topic.FileScanFileRequest{
					FileId: id,
					Deep:   true,
				}

				body, _ := json.Marshal(taskReq)
				if err = h.taskEngine.PushMessage(
					ctx.GetContext().
						WithValue(consts.CtxKeyFullPath, fullPath).
						WithValue(consts.CtxKeyInvokeHandlerName, "入库执行器"),
					taskReq.Topic(), body); err != nil {
					logger.Error("下发文件扫描任务失败", zap.Error(err))
				}

				if _, err = h.authIngestLogService.Create(ctx.GetContext(),
					req.PlanId, autoingest.LogLevelInfo,
					fmt.Sprintf("新增入库：%s", fullPath),
				); err != nil {
					logger.Error("创建入库日志失败", zap.Error(err))
				}
			}
		}

		return nil
	}
}
