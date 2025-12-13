package storage

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	mountPointSvi "github.com/xxcheng123/cloudpan189-share/internal/services/mountpoint"
	"go.uber.org/zap"
)

type batchCreateTextRequest struct {
	Content           string `json:"content" binding:"required"`
	CloudToken        int64  `json:"cloudToken" binding:"required"`
	EnableAutoRefresh bool   `json:"enableAutoRefresh"`
	RefreshInterval   int    `json:"refreshInterval"`
	ShareAccessCode   string `json:"shareAccessCode"`
}

type batchCreateTextResponse struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

var (
	reFolderID  = regexp.MustCompile(`^\d+$`)
	reShareLink = regexp.MustCompile(`cloud\.189\.cn\/t\/([a-zA-Z0-9]+)`)
)

func (h *handler) BatchCreateFromText() httpcontext.HandlerFunc {
	return func(ctx *httpcontext.Context) {
		req := new(batchCreateTextRequest)
		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.AbortWithInvalidParams(err)
			return
		}

		tokenInfo, err := h.cloudTokenService.Query(ctx.GetContext(), req.CloudToken)
		if err != nil || tokenInfo == nil {
			ctx.Fail(busCodeStorageCloudTokenNotExist)
			return
		}

		if req.RefreshInterval <= 0 {
			req.RefreshInterval = 30
		}

		lines := strings.Split(req.Content, "\n")
		successCount := 0
		failCount := 0

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parts := strings.Fields(line)
			resourceStr := parts[0]
			// 这里的 AccessCode 解析虽然保留，但目前的 CreateRequest 结构体中没有对应字段
			// 如果业务需要存储访问码，请在 mountPointSvi.CreateRequest 中添加字段，或将其编码进 FullPath 中
			// currentAccessCode := req.ShareAccessCode
			// if len(parts) > 1 {
			// 	currentAccessCode = parts[1]
			// }

			var createReq *mountPointSvi.CreateRequest

			// 匹配分享链接
			if matches := reShareLink.FindStringSubmatch(resourceStr); len(matches) > 1 {
				shareCode := matches[1]
				// Service 层逻辑：name = parts[len(parts)-1]，所以构造 FullPath 来传递 Name
				name := "分享_" + shareCode

				createReq = &mountPointSvi.CreateRequest{
					TokenId:           req.CloudToken,
					FullPath:          "/" + name, // 构造路径以便 Service 提取名称
					OsType:            protocolSubscribeShare,
					EnableAutoRefresh: req.EnableAutoRefresh,
					RefreshInterval:   req.RefreshInterval,
					FileId: 0,
				}

			// 匹配纯数字文件夹ID
			} else if reFolderID.MatchString(resourceStr) {
				fileId, _ := strconv.ParseInt(resourceStr, 10, 64)
				name := "文件夹_" + resourceStr

				createReq = &mountPointSvi.CreateRequest{
					TokenId:           req.CloudToken,
					FullPath:          "/" + name, // 构造路径以便 Service 提取名称
					OsType:            protocolSubscribe,
					FileId:            fileId,
					EnableAutoRefresh: req.EnableAutoRefresh,
					RefreshInterval:   req.RefreshInterval,
				}
			} else {
				ctx.GetContext().Warn("无法识别的资源格式", zap.String("line", line))
				failCount++
				continue
			}

			// 执行创建
			if createReq != nil {
				_, err := h.mountPointService.Create(ctx.GetContext(), createReq)
				if err != nil {
					ctx.GetContext().Error("批量导入创建失败",
						zap.Error(err),
						zap.String("line", line),
					)
					failCount++
				} else {
					successCount++
				}
			}
		}

		ctx.Success(&batchCreateTextResponse{
			Total:   len(lines),
			Success: successCount,
			Failed:  failCount,
		})
	}
}
