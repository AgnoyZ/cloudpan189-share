package advance

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
)

type getShareInfoRequest struct {
	ShareCode       string `form:"shareCode" binding:"required" example:"abc123"`
	ShareAccessCode string `form:"shareAccessCode" example:"1234"`
}

type _ = cloudbridge.ShareInfo

// GetShareInfo 获取分享信息
// @Summary 获取分享信息
// @Description 根据分享码获取分享的详细信息，包括文件名、是否为文件夹、分享时间等
// @Tags 存储高级功能
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param shareCode query string true "分享码" example("abc123")
// @Param shareAccessCode query string false "分享访问码" example("1234")
// @Success 200 {object} httpcontext.Response{data=cloudbridge.ShareInfo} "获取分享信息成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "获取分享详情失败，code=8005"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/storage/advance/share_info [get]
func (h *handler) GetShareInfo() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(getShareInfoRequest)
		if err := ctx.ShouldBindQuery(req); err != nil {
			ctx.AbortWithInvalidParams(err)

			return
		}

		shareInfo, err := h.cloudBridgeService.GetShareInfo(ctx.GetContext(), req.ShareCode, req.ShareAccessCode)
		if err != nil {
			ctx.Fail(codeStorageAdvanceGetShareInfoError.WithError(err))

			return
		}

		ctx.Success(shareInfo)
	}
}
