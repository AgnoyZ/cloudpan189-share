package main

import (
	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/consumer"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/scheduler"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/shutdown"
	"go.uber.org/zap"
)

func main() {
	svc, err := bootstrap.New(configs.Get())
	if err != nil {
		panic(err)
	}

	defer svc.Close()

	logger := svc.GetLogger("main")
	taskEngine := svc.GetTaskEngine()

	if err = consumer.Start(svc); err != nil {
		panic(err)
	}

	var (
		closeBar func()
	)

	// 启动scheduler
	if closeBar, err = scheduler.Start(svc); err != nil {
		panic(err)
	}

	// 启动HTTP服务在单独的goroutine中，避免阻塞
	go func() {
		if err = http.Start(svc); err != nil {
			logger.Error("http server failed", zap.Error(err))
		}
	}()

	logger.Info("system running....")

	shutdown.Close(func() {
		logger.Info("close shutdown")

		if err = taskEngine.Stop(); err != nil {
			logger.Error("close task engine failed", zap.Error(err))
		}

		closeBar()

		logger.Info("close successful, bye~")
	})
}
