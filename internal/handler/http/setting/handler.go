package setting

import (
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"

	settingSvi "github.com/xxcheng123/cloudpan189-share/internal/services/setting"
	userSvi "github.com/xxcheng123/cloudpan189-share/internal/services/user"
)

type Handler interface {
	InitSystem() httpcontext.HandlerFunc
	Info() httpcontext.HandlerFunc
}

var bi = httpcontext.NewBusinessGenerator(consts.BusCodeSettingStartCode)

var (
	codeInitSettingErr   = bi.Next("初始化系统时发生错误")
	codeInitSuperUserErr = bi.Next("初始化超级管理员时发生错误")
	codeQueryFailed      = bi.Next("查询系统配置失败")
)

type handler struct {
	userService    userSvi.Service
	settingService settingSvi.Service
	initTime       time.Time
}

func NewHandler(userService userSvi.Service, settingService settingSvi.Service) Handler {
	return &handler{
		userService:    userService,
		settingService: settingService,
		initTime:       time.Now(),
	}
}
