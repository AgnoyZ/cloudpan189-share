package http

import (
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/taskstate"

	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/cloudtoken"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/file"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/setting"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/storage"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/storage/advance"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/usergroup"

	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/handler/http/user"

	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
	cloudtokenSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
	filetasklogSvi "github.com/xxcheng123/cloudpan189-share/internal/services/filetasklog"
	group2fileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/group2file"
	mountPointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"
	settingSvi "github.com/xxcheng123/cloudpan189-share/internal/services/setting"
	userSvi "github.com/xxcheng123/cloudpan189-share/internal/services/user"
	userGroupSvi "github.com/xxcheng123/cloudpan189-share/internal/services/usergroup"
	verifySvi "github.com/xxcheng123/cloudpan189-share/internal/services/verify"
	virtualfileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
)

func Start(svc bootstrap.ServiceContext) {
	const (
		handlerName = "http"
	)

	var (
		engine = svc.GetHTTPEngine()

		logger     = svc.GetLogger(handlerName)
		taskEngine = svc.GetTaskEngine()

		wrapper = httpcontext.NewHandlerFuncWrapper(logger)
		wrap    = wrapper.Wrap
	)

	var (
		userService        = userSvi.NewService(svc)
		userGroupService   = userGroupSvi.NewService(svc)
		group2FileService  = group2fileSvi.NewService(svc)
		settingService     = settingSvi.NewService(svc)
		virtualFileService = virtualfileSvi.NewService(svc)
		cloudBridgeService = cloudbridgeSvi.NewService(svc)
		cloudTokenService  = cloudtokenSvi.NewService(svc)
		mountPointService  = mountPointSvi.NewService(svc)
		fileTaskLogService = filetasklogSvi.NewService(svc)
		verifyService      = verifySvi.NewService(svc)
	)

	var (
		userHandler           = user.NewHandler(userService, userGroupService)
		settingHandler        = setting.NewHandler(userService, settingService)
		userGroupHandler      = usergroup.NewHandler(userGroupService, group2FileService, userService)
		storageHandler        = storage.NewHandler(taskEngine, virtualFileService, cloudBridgeService, cloudTokenService, mountPointService, fileTaskLogService)
		storageAdvanceHandler = advance.NewHandler(cloudBridgeService, cloudTokenService)
		cloudTokenHandler     = cloudtoken.NewHandler(cloudTokenService, mountPointService)
		fileHandler           = file.NewHandler(virtualFileService, verifyService, cloudTokenService, cloudBridgeService, mountPointService)
		taskStateHandler      = taskstate.NewHandler(taskEngine, fileTaskLogService)
	)

	var (
		userMiddleware = newAuthMiddleware(userService)
	)

	openapiRouter := engine.Group("/api", httpcontext.LoggerHandler(logger))

	{
		userRouter := openapiRouter.Group("/user")
		{
			userRouter.POST("/login", wrap(userHandler.Login()))
			userRouter.POST("/refresh_token", wrap(userHandler.RefreshToken()))
		}

		userRouterWithAdminAuth := openapiRouter.Group("/user", wrap(userMiddleware.Auth(true)))
		{
			userRouterWithAdminAuth.POST("/add", wrap(userHandler.Add()))
			userRouterWithAdminAuth.POST("/del", wrap(userHandler.Del()))
			userRouterWithAdminAuth.POST("/update", wrap(userHandler.Update()))
			userRouterWithAdminAuth.POST("/toggle_status", wrap(userHandler.ToggleStatus()))
			userRouterWithAdminAuth.GET("/list", wrap(userHandler.List()))
			userRouterWithAdminAuth.POST("/modify_pass", wrap(userHandler.ModifyPass()))
			userRouterWithAdminAuth.POST("/bind_group", wrap(userHandler.BindGroup()))
		}

		userRouterWithBaseAuth := openapiRouter.Group("/user", wrap(userMiddleware.Auth()))
		{
			userRouterWithBaseAuth.GET("/info", wrap(userHandler.Info()))
			userRouterWithBaseAuth.POST("/modify_own_pass", wrap(userHandler.ModifyOwnPass()))
		}
	}

	{
		userGroupRouter := openapiRouter.Group("/user_group", wrap(userMiddleware.Auth(true)))
		{
			userGroupRouter.POST("/add", wrap(userGroupHandler.Add()))
			userGroupRouter.POST("/delete", wrap(userGroupHandler.Delete()))
			userGroupRouter.POST("/modify_name", wrap(userGroupHandler.ModifyName()))
			userGroupRouter.GET("/list", wrap(userGroupHandler.List()))
			userGroupRouter.POST("/batch_bind_files", wrap(userGroupHandler.BatchBindFiles()))
			userGroupRouter.GET("/bind_files", wrap(userGroupHandler.GetBindFiles()))
		}
	}

	{
		storageRouter := openapiRouter.Group("/storage", wrap(userMiddleware.Auth()))
		{
			storageRouter.POST("/add", wrap(storageHandler.Add()))
			storageRouter.POST("/delete", wrap(storageHandler.Delete()))
			storageRouter.GET("/list", wrap(storageHandler.List()))
			storageRouter.POST("/refresh", wrap(storageHandler.Refresh()))
			storageRouter.POST("/toggle_auto_refresh", wrap(storageHandler.ToggleAutoRefresh()))
			storageRouter.POST("/modify_token", wrap(storageHandler.ModifyToken()))
		}

		storageAdvanceRouter := openapiRouter.Group("/storage/advance", wrap(userMiddleware.Auth()))
		{
			storageAdvanceRouter.GET("/person/files", wrap(storageAdvanceHandler.GetPersonFiles()))
			storageAdvanceRouter.GET("/family/files", wrap(storageAdvanceHandler.GetFamilyFiles()))
			storageAdvanceRouter.GET("/family/list", wrap(storageAdvanceHandler.FamilyList()))
			storageAdvanceRouter.GET("/get_subscribe_user", wrap(storageAdvanceHandler.GetSubscribeUser()))
			storageAdvanceRouter.GET("/share_info", wrap(storageAdvanceHandler.GetShareInfo()))
		}
	}

	{
		fileRouter := openapiRouter.Group("/file", wrap(userMiddleware.Auth()))
		{
			fileRouter.GET("/search", wrap(fileHandler.Search()))
			fileRouter.POST("/create_download_url", wrap(fileHandler.CreateDownloadURL()))
			fileRouter.GET("/open/*fullPath", wrap(fileHandler.Open()))
		}

		{
			openapiRouter.GET("/file/download/:fileId", wrap(fileHandler.Download()))
		}
	}

	{
		cloudTokenRouter := openapiRouter.Group("/cloud_token", wrap(userMiddleware.Auth(true)))
		{
			cloudTokenRouter.POST("/init_qrcode", wrap(cloudTokenHandler.InitQrcode()))
			cloudTokenRouter.POST("/check_qrcode", wrap(cloudTokenHandler.CheckQrcode()))
			cloudTokenRouter.POST("/username_login", wrap(cloudTokenHandler.UsernameLogin()))
			cloudTokenRouter.POST("/modify_name", wrap(cloudTokenHandler.ModifyName()))
			cloudTokenRouter.POST("/delete", wrap(cloudTokenHandler.Delete()))
			cloudTokenRouter.GET("/list", wrap(cloudTokenHandler.List()))
			cloudTokenRouter.GET("/:id", wrap(cloudTokenHandler.Query()))
		}
	}

	{
		taskStateRouter := openapiRouter.Group("/task_state", wrap(userMiddleware.Auth(true)))
		{
			taskStateRouter.GET("/file_log/list", wrap(taskStateHandler.FileLogList()))
			taskStateRouter.GET("/task_engine/list", wrap(taskStateHandler.TaskEngineList()))
		}
	}

	{
		openapiRouter.POST("/setting/init_system", wrap(settingHandler.InitSystem()))
		openapiRouter.GET("/setting/info", wrap(settingHandler.Info()))
	}
	{
		settingBaseRouter := openapiRouter.Group("/setting", wrap(userMiddleware.Auth()))
		{
			settingBaseRouter.GET("/addition", wrap(settingHandler.Addition()))
		}
	}

	{
		settingAdminRouter := openapiRouter.Group("/setting", wrap(userMiddleware.Auth(true)))
		{
			settingAdminRouter.POST("/modify_title", wrap(settingHandler.ModifyTitle()))
			settingAdminRouter.POST("/modify_base_url", wrap(settingHandler.ModifyBaseURL()))
			settingAdminRouter.POST("/toggle_enable_auth", wrap(settingHandler.ToggleEnableAuth()))
			settingAdminRouter.POST("/modify_addition", wrap(settingHandler.ModifyAddition()))
		}
	}
}
