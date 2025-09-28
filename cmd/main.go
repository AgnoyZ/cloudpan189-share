package main

import (
	"fmt"

	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/consumer"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/dav"
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
	httpEngine := svc.GetHTTPEngine()
	port := svc.GetPort()

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

	http.Start(svc)
	dav.Start(svc)

	go func() {
		if err = httpEngine.Run(fmt.Sprintf(":%d", port)); err != nil {
			panic(err)
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
