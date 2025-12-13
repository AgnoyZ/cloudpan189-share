package storage

import (
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"
)

// BatchParseFromText 批量解析文本（获取真实名称和ID，不创建）
func (h *handler) BatchParseFromText() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		// 使用 topic 包中定义的请求结构
		req := new(topic.BatchParseTextRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		// 校验 CloudToken 是否存在
		tokenInfo, err := h.cloudTokenService.Query(ctx.GetContext(), req.CloudToken)
		if err != nil || tokenInfo == nil {
			ctx.Fail(busCodeStorageCloudTokenNotExist)
			return
		}

		// 调用 Service 层进行解析 (核心逻辑在 Service 中)
		result, err := h.mountPointService.BatchParseText(ctx.GetContext(), req)
		if err != nil {
			ctx.Error(err)
			return
		}

		// 返回解析结果列表
		ctx.Success(result)
	}
}
