package media

import (
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	mediaconfigSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mediaconfig"
	mediafileSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mediafile"
)

// Handler 定义 media 相关的 HTTP 处理器接口
type Handler interface {
	// 媒体配置
	ConfigInit() httpcontext.HandlerFunc
	ConfigInfo() httpcontext.HandlerFunc
	ConfigUpdate() httpcontext.HandlerFunc
	ConfigToggle() httpcontext.HandlerFunc
}

var bi = httpcontext.NewBusinessGenerator(consts.BusCodeMediaStartCode)

var (
	// 配置相关错误码
	codeConfigQueryFailed  = bi.Next("查询媒体配置失败")
	codeConfigInitFailed   = bi.Next("初始化媒体配置失败")
	codeConfigUpdateFailed = bi.Next("更新媒体配置失败")
	codeConfigToggleFailed = bi.Next("切换媒体配置启用状态失败")
)

// handler 依赖的服务
type handler struct {
	mediaConfigService mediaconfigSvi.Service
	mediaFileService   mediafileSvi.Service
}

// NewHandler 构造函数
func NewHandler(
	mediaConfigService mediaconfigSvi.Service,
	mediaFileService mediafileSvi.Service,
) Handler {
	return &handler{
		mediaConfigService: mediaConfigService,
		mediaFileService:   mediaFileService,
	}
}
