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

		// 校验 CloudToken
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
			currentAccessCode := req.ShareAccessCode
			if len(parts) > 1 {
				currentAccessCode = parts[1]
			}

			// 注意：这里使用 mountPointSvi.CreateRequest 而不是 models.MountPoint
			var createReq *mountPointSvi.CreateRequest

			if matches := reShareLink.FindStringSubmatch(resourceStr); len(matches) > 1 {
				shareCode := matches[1]
				// 分享链接
				createReq = &mountPointSvi.CreateRequest{
					Name:              "分享导入_" + shareCode,
					TokenId:           req.CloudToken,
					OsType:            protocolSubscribeShare,
					EnableAutoRefresh: req.EnableAutoRefresh,
					RefreshInterval:   req.RefreshInterval,
					// 假设 CreateRequest 结构体中有 AccessKey 或 Remark 来存 ShareCode
					// 如果编译报错说没有 AccessKey，请尝试用 Remark 或 ShareId
					AccessKey: shareCode,
					Password:  currentAccessCode,
				}
			} else if reFolderID.MatchString(resourceStr) {
				// 文件夹ID
				fileId, _ := strconv.ParseInt(resourceStr, 10, 64)
				createReq = &mountPointSvi.CreateRequest{
					Name:              "文件夹导入_" + resourceStr,
					TokenId:           req.CloudToken,
					OsType:            protocolSubscribe,
					EnableAutoRefresh: req.EnableAutoRefresh,
					RefreshInterval:   req.RefreshInterval,
					FileId:            fileId,
				}
			} else {
				failCount++
				continue
			}

			// 调用 Create 方法，传入 CreateRequest
			if _, err := h.mountPointService.Create(ctx.GetContext(), createReq); err != nil {
				h.taskEngine.GetLogger().Error("批量创建失败", zap.String("source", line), zap.Error(err))
				failCount++
			} else {
				successCount++
			}
		}

		ctx.Success(&batchCreateTextResponse{
			Total:   len(lines),
			Success: successCount,
			Failed:  failCount,
		})
	}
}
