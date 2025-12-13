package advance

import (
	"regexp"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
)

type getShareInfoRequest struct {
	// 允许前端传 code 或者完整 url
	ShareCode       string `form:"shareCode" binding:"required" example:"abc12345"`
	ShareAccessCode string `form:"shareAccessCode" example:"1234"`
}

// 预编译正则，提取 /t/ 后面的字符
var reShareLink = regexp.MustCompile(`cloud\.189\.cn\/t\/([a-zA-Z0-9]+)`)

// GetShareInfo 获取分享信息
// @Summary 获取分享信息
// @Description 根据分享码获取分享的详细信息，支持直接传入完整分享链接
// @Tags 存储高级功能
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param shareCode query string true "分享码或完整链接" example("https://cloud.189.cn/t/abc12345")
// @Param shareAccessCode query string false "分享访问码" example("1234")
// @Success 200 {object} httpcontext.Response{data=cloudbridge.ShareInfo} "获取分享信息成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败"
// @Failure 400 {object} httpcontext.Response "获取分享详情失败"
// @Router /api/storage/advance/share_info [get]
func (h *handler) GetShareInfo() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(getShareInfoRequest)
		if err := ctx.ShouldBindQuery(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		// 如果前端传的是 https://cloud.189.cn/t/xxx 形式的链接，则提取 xxx 作为分享码
		if matches := reShareLink.FindStringSubmatch(req.ShareCode); len(matches) > 1 {
			req.ShareCode = matches[1]
		}

		shareInfo, err := h.cloudBridgeService.GetShareInfo(ctx.GetContext(), req.ShareCode, req.ShareAccessCode)
		if err != nil {
			ctx.Fail(codeStorageAdvanceGetShareInfoError.WithError(err))
			return
		}

		ctx.Success(shareInfo)
	}
}
