package scheduler

import (
	"encoding/json"
	"runtime/debug"
	"sync"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/taskengine"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	"github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	"github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"
	"github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
	"go.uber.org/zap"
)

type RefreshFileScheduler struct {
	running            bool
	mu                 sync.Mutex
	ctx                context.Context
	cancel             context.CancelFunc
	mountPointService  mountpoint.Service
	fileTaskLogService filetasklog.Service
	virtualFileService virtualfile.Service
	taskEngine         taskengine.TaskEngine
	cleanupMu          sync.Mutex
	lastCleanupAt      time.Time
}

func NewRefreshFileScheduler(
	mountPointService mountpoint.Service,
	fileTaskLogService filetasklog.Service,
	virtualFileService virtualfile.Service,
	taskEngine taskengine.TaskEngine,
) Scheduler {
	return &RefreshFileScheduler{
		mountPointService:  mountPointService,
		fileTaskLogService: fileTaskLogService,
		virtualFileService: virtualFileService,
		taskEngine:         taskEngine,
		running:            false,
	}
}

func (s *RefreshFileScheduler) Start(ctx context.Context) error {
	if !s.mu.TryLock() {
		return ErrSchedulerRunning
	}
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.ctx, s.cancel = context.WithCancel(ctx)

	s.running = true

	gopool.Go(func() {
		for s.doJob() {
		}

		ctx.Info("文件刷新执行器已停止~")
	})

	return nil
}

func (s *RefreshFileScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.cancel()
	s.running = false
}

func (s *RefreshFileScheduler) doJob() bool {
	ctx := s.ctx

	defer func() {
		if r := recover(); r != nil {
			ctx.Error("文件刷新执行器发生异常",
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())))
		}
	}()

	select {
	case <-ctx.Done():
		ctx.Info("文件刷新执行器停止")

		return false
	case <-time.After(time.Minute):
		s.cleanupZeroFileMountPoints(ctx)

		mountPoints, err := s.mountPointService.GetAutoRefreshList(ctx, &mountpoint.GetAutoRefreshListRequest{})
		if err != nil {
			ctx.Error("查询挂载点失败", zap.Error(err))

			return true
		}

		ctx.Debug("文件刷新执行器查询到挂载点数量", zap.Int("count", len(mountPoints)))

		now := time.Now()

		// 获取今天零点
		midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		// 计算从零点到现在的分钟数
		minutesSinceMidnight := int(now.Sub(midnight).Minutes())

		for _, mp := range mountPoints {
			// 只处理启用自动刷新的挂载点
			if !mp.EnableAutoRefresh {
				continue
			}

			if minutesSinceMidnight%mp.RefreshInterval != 0 {
				continue
			}

			ctx.Info("文件扫描执行器触发",
				zap.Int64("mount_point_id", mp.ID),
				zap.Int64("file_id", mp.FileId),
				zap.String("full_path", mp.FullPath),
				zap.Int("refresh_interval", mp.RefreshInterval))

			// 创建文件扫描任务
			taskReq := &topic.FileScanFileRequest{
				FileId: mp.FileId,
				Deep:   mp.EnableDeepRefresh,
			}

			body, _ := json.Marshal(taskReq)
			taskCtx := ctx.
				WithValue(consts.CtxKeyFullPath, mp.FullPath).
				WithValue(consts.CtxKeyInvokeHandlerName, "定时任务")

			if err = s.taskEngine.PushMessage(taskCtx, taskReq.Topic(), body); err != nil {
				ctx.Error("推送文件扫描任务失败",
					zap.Int64("mount_point_id", mp.ID),
					zap.Int64("file_id", mp.FileId),
					zap.String("full_path", mp.FullPath),
					zap.Error(err))
			} else {
				ctx.Info("下发文件扫描任务成功",
					zap.Int64("mount_point_id", mp.ID),
					zap.Int64("file_id", mp.FileId),
					zap.String("full_path", mp.FullPath))
			}
		}
	}

	return true
}

func (s *RefreshFileScheduler) cleanupZeroFileMountPoints(ctx context.Context) {
	s.cleanupMu.Lock()
	defer s.cleanupMu.Unlock()

	cleanupInterval := time.Duration(shared.SettingAddition.ZeroFileCleanupInterval) * time.Minute
	if cleanupInterval <= 0 {
		cleanupInterval = time.Hour
	}

	if !s.lastCleanupAt.IsZero() && time.Since(s.lastCleanupAt) < cleanupInterval {
		return
	}

	s.lastCleanupAt = time.Now()

	enableAutoRefresh := true
	mountPoints, err := s.mountPointService.List(ctx, &mountpoint.ListRequest{
		EnableAutoRefresh: &enableAutoRefresh,
		NoPaginate:        true,
	})
	if err != nil {
		ctx.Error("查询零文件清理候选挂载点失败", zap.Error(err))
		return
	}

	if len(mountPoints) == 0 {
		return
	}

	fileIdList := make([]int64, 0, len(mountPoints))
	for _, mp := range mountPoints {
		fileIdList = append(fileIdList, mp.FileId)
	}

	fileCountList, err := s.virtualFileService.GroupCountByTopId(ctx, &virtualfile.GroupCountByTopIdRequest{})
	if err != nil {
		ctx.Error("查询挂载点文件数量失败", zap.Error(err))
		return
	}

	fileCountMap := make(map[int64]int64, len(fileCountList))
	for _, item := range fileCountList {
		fileCountMap[item.TopId] = item.Count
	}

	latestScanStatusMap, err := s.fileTaskLogService.LatestStatusByFileIDs(ctx, topic.FileScanFileRequest{}.Topic().String(), fileIdList)
	if err != nil {
		ctx.Error("查询挂载点扫描日志失败", zap.Error(err))
		return
	}

	zeroIDs := make([]int64, 0)
	for _, mp := range mountPoints {
		if !mp.ShouldAutoDeleteWhenZeroFiles() {
			continue
		}
		if fileCountMap[mp.FileId] != 0 {
			continue
		}
		if latestScanStatusMap[mp.FileId] != models.StatusCompleted {
			ctx.Debug("跳过未完成首次成功扫描的零文件挂载点",
				zap.Int64("file_id", mp.FileId),
				zap.String("latest_scan_status", latestScanStatusMap[mp.FileId]))
			continue
		}
		zeroIDs = append(zeroIDs, mp.FileId)
	}

	if len(zeroIDs) == 0 {
		return
	}

	taskReq := &topic.FileBatchDeleteRequest{IDs: zeroIDs}
	body, _ := json.Marshal(taskReq)
	taskCtx := ctx.WithValue(consts.CtxKeyInvokeHandlerName, "零文件定期清理")
	if err = s.taskEngine.PushMessage(taskCtx, taskReq.Topic(), body); err != nil {
		ctx.Error("推送零文件挂载点清理任务失败", zap.Int64s("ids", zeroIDs), zap.Error(err))
		return
	}

	ctx.Info("已推送零文件挂载点清理任务", zap.Int64s("ids", zeroIDs), zap.Int("count", len(zeroIDs)))
}
