package consumer

import (
	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/taskcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/consumer/file"
	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
	cloudtokenSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
	filetasklogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	mountPointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"
	virtualfileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
)

func Start(svc bootstrap.ServiceContext) error {
	var (
		handlerName = "consumer"
		logger      = svc.GetLogger(handlerName)

		wrapper = taskcontext.NewHandlerFuncWrapper(logger)
		wrap    = wrapper.Wrap

		taskEngine = svc.GetTaskEngine()
	)

	var (
		virtualFileService = virtualfileSvi.NewService(svc)
		cloudBridgeService = cloudbridgeSvi.NewService(svc)
		cloudTokenService  = cloudtokenSvi.NewService(svc)
		mountPointService  = mountPointSvi.NewService(svc)
		fileTaskLogService = filetasklogSvi.NewService(svc)
	)

	var (
		fileHandler = file.NewHandler(virtualFileService, cloudBridgeService, cloudTokenService, mountPointService, fileTaskLogService)
	)

	{
		if err := taskEngine.RegisterProcessor(new(topic.FileScanFileRequest).Topic(), wrap(fileHandler.ScanFile())); err != nil {
			logger.Error("注册文件扫描处理器失败")

			return err
		}

		if err := taskEngine.RegisterProcessor(new(topic.FileClearFileRequest).Topic(), wrap(fileHandler.ClearFile())); err != nil {
			logger.Error("注册文件清理处理器失败")

			return err
		}
	}

	logger.Info("consumer handler start")

	return taskEngine.Start()
}
