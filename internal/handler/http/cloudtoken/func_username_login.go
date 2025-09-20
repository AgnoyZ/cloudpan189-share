package cloudtoken

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
)

type (
	usernameLoginRequest  = cloudtoken.UsernameLoginRequest
	UsernameLoginResponse = cloudtoken.UsernameLoginResponse
)

// UsernameLogin 用户名密码登录
// @Summary 用户名密码登录
// @Description 使用用户名和密码登录云盘，创建或更新云盘令牌
// @Tags 云盘令牌管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body usernameLoginRequest true "用户名登录请求"
// @Success 200 {object} httpcontext.Response{data=cloudtoken.UsernameLoginResponse} "登录成功"
// @Failure 400 {object} httpcontext.Response "参数验证失败，code=99998"
// @Failure 400 {object} httpcontext.Response "用户名登录失败，code=5006"
// @Failure 401 {object} httpcontext.Response "未授权访问"
// @Failure 403 {object} httpcontext.Response "权限不足"
// @Router /api/cloud_token/username_login [post]
func (h *handler) UsernameLogin() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(usernameLoginRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)

			return
		}

		resp, err := h.cloudTokenService.UsernameLogin(ctx.GetContext(), req)
		if err != nil {
			ctx.Fail(codeUsernameLoginFailed.WithError(err))

			return
		}

		ctx.Success(resp)
	}
}
