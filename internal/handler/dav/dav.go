package dav

import (
	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"

	userSvi "github.com/xxcheng123/cloudpan189-share/internal/services/user"
	verifySvi "github.com/xxcheng123/cloudpan189-share/internal/services/verify"
	virtualfileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/virtualfile"
)

var (
	davMethods = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS", "PROPFIND", "MKCOL", "MOVE", "LOCK", "UNLOCK"}
)

func Start(svc bootstrap.ServiceContext) {
	const (
		handlerName = "dav"
	)

	var (
		engine  = svc.GetHTTPEngine()
		logger  = svc.GetLogger(handlerName)
		wrapper = httpcontext.NewHandlerFuncWrapper(logger)
		wrap    = wrapper.Wrap
	)

	var (
		userService        = userSvi.NewService(svc)
		virtualFileService = virtualfileSvi.NewService(svc)
		verifyService      = verifySvi.NewService(svc)
	)

	davRouter := engine.Group("/dav",
		wrap(newBalanceMiddleware().Balance()),
		wrap(newAuthMiddleware(userService).Auth()),
	)

	workEngine := &workEngine{
		verifyService:      verifyService,
		virtualFileService: virtualFileService,
	}

	for _, method := range davMethods {
		davRouter.Handle(method, "/*path", wrap(workEngine.Open()))
		davRouter.Handle(method, "", wrap(workEngine.Open()))
	}
}
