package scheduler

import (
	errors2 "errors"

	"github.com/pkg/errors"

	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"

	filetasklogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	mountpointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"

	stdContext "context"
)

type Scheduler interface {
	Start(ctx context.Context) error
	Stop()
}

var (
	ErrSchedulerRunning = errors.New("scheduler is running")
)

func Start(svc bootstrap.ServiceContext) (func(), error) {
	const (
		handlerName = "scheduler"
	)

	var (
		logger = svc.GetLogger(handlerName)

		errs []error

		ctx = context.NewContext(stdContext.Background(), context.WithLogger(logger))
	)

	var (
		fileTaskLogService = filetasklogSvi.NewService(svc)
		mountPointService  = mountpointSvi.NewService(svc)
		taskEngine         = svc.GetTaskEngine()
	)

	fileTaskLogCheckScheduler := NewFileTaskLogCheckScheduler(fileTaskLogService)
	if err := fileTaskLogCheckScheduler.Start(ctx); err != nil {
		errs = append(errs, err)
	}

	refreshFileScheduler := NewRefreshFileScheduler(mountPointService, taskEngine)
	if err := refreshFileScheduler.Start(ctx); err != nil {
		errs = append(errs, err)
	}

	schedulers := []Scheduler{
		fileTaskLogCheckScheduler,
		refreshFileScheduler,
	}

	return closeBar(schedulers), errors2.Join(errs...)
}

func closeBar(schedulers []Scheduler) func() {
	return func() {
		for _, scheduler := range schedulers {
			scheduler.Stop()
		}
	}
}
